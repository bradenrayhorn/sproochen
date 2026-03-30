package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const apiBaseURL = "https://lod.lu/api/en/entry"

type apiResponse struct {
	Entry entry `json:"entry"`
}

type entry struct {
	LodID           string           `json:"lod_id"`
	PartOfSpeech    string           `json:"partOfSpeech"`
	PartOfSpeechLbl string           `json:"partOfSpeechLabel"`
	Lemma           string           `json:"lemma"`
	NRuleForm       string           `json:"nRuleForm"`
	Tables          entryTables      `json:"tables"`
	AudioFiles      audioFiles       `json:"audioFiles"`
	MicroStructures []microStructure `json:"microStructures"`
}

type entryTables struct {
	VerbConjugation *verbConjugation `json:"verbConjugation,omitempty"`
}

type verbConjugation struct {
	Attributes     verbConjugationAttributes `json:"@attributes"`
	Infinitive     string                    `json:"infinitive"`
	PastParticiple stringOrStrings           `json:"pastParticiple"`
	AuxiliaryVerb  stringOrStrings           `json:"auxiliaryVerb"`
	Indicative     verbTenseGroup            `json:"indicative"`
	Conditional    verbTenseGroup            `json:"conditional"`
	Imperative     verbTenseGroup            `json:"imperative"`
}

type stringOrStrings []string

func (s *stringOrStrings) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		trimmed := strings.TrimSpace(single)
		if trimmed == "" {
			*s = nil
		} else {
			*s = []string{trimmed}
		}
		return nil
	}

	var many []string
	if err := json.Unmarshal(data, &many); err == nil {
		*s = dedupeStrings(many)
		return nil
	}

	return fmt.Errorf("unsupported stringOrStrings value: %s", string(data))
}

type verbConjugationAttributes struct {
	ID            string `json:"id"`
	Model         string `json:"model"`
	SeparableVerb string `json:"separableVerb"`
}

type verbTenseGroup struct {
	Present        stringMap `json:"present,omitempty"`
	PresentPerfect stringMap `json:"presentPerfect,omitempty"`
	PastPerfect    stringMap `json:"pastPerfect,omitempty"`
}

func (g *verbTenseGroup) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" || trimmed == "[]" {
		*g = verbTenseGroup{}
		return nil
	}

	type alias verbTenseGroup
	var tmp alias
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	*g = verbTenseGroup(tmp)
	return nil
}

type stringMap map[string]string

func (m *stringMap) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" || trimmed == "[]" {
		*m = nil
		return nil
	}

	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	cleaned := make(map[string]string, len(raw))
	for k, v := range raw {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		cleaned[k] = v
	}
	if len(cleaned) == 0 {
		*m = nil
	} else {
		*m = cleaned
	}
	return nil
}

type audioFiles struct {
	OGG string `json:"ogg"`
}

type microStructure struct {
	AuxiliaryVerb    string            `json:"auxiliaryVerb"`
	PastParticiple   []string          `json:"pastParticiple"`
	GrammaticalUnits []grammaticalUnit `json:"grammaticalUnits"`
}

type grammaticalUnit struct {
	Meanings []meaning `json:"meanings"`
}

type meaning struct {
	MeaningID         string                  `json:"meaningID"`
	Number            int                     `json:"number"`
	SecondaryHeadword string                  `json:"secondaryHeadword"`
	TargetLanguages   map[string]languageData `json:"targetLanguages"`
	Examples          []example               `json:"examples"`
}

type languageData struct {
	Parts []languagePart `json:"parts"`
}

type languagePart struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type example struct {
	Parts []examplePart `json:"parts"`
}

type examplePart struct {
	Type  string         `json:"type"`
	Parts []exampleToken `json:"parts"`
}

type exampleToken struct {
	Type                 string `json:"type"`
	Content              string `json:"content"`
	JoinWithPreviousWord bool   `json:"joinWithPreviousWord"`
}

type flashcard struct {
	LodID            string     `json:"lod_id"`
	MeaningID        string     `json:"meaning_id,omitempty"`
	MeaningNumber    int        `json:"meaning_number,omitempty"`
	NativeLanguage   string     `json:"native_language"`
	TargetLanguage   string     `json:"target_language"`
	PartOfSpeech     string     `json:"part_of_speech"`
	ExampleSentences []string   `json:"example_sentences,omitempty"`
	EnglishClarifier string     `json:"english_clarifier,omitempty"`
	VerbForms        *verbForms `json:"verb_forms,omitempty"`
	EntryAudioFile   string     `json:"entry_audio_file,omitempty"`
	CachedResponse   string     `json:"cached_response,omitempty"`
}

type verbForms struct {
	NRuleForm        string            `json:"n_rule_form,omitempty"`
	Infinitive       string            `json:"infinitive,omitempty"`
	PastParticiples  []string          `json:"past_participles,omitempty"`
	AuxiliaryVerbs   []string          `json:"auxiliary_verbs,omitempty"`
	ConjugationID    string            `json:"conjugation_id,omitempty"`
	ConjugationModel string            `json:"conjugation_model,omitempty"`
	SeparableVerb    string            `json:"separable_verb,omitempty"`
	Indicative       *verbTenseGroup   `json:"indicative,omitempty"`
	Conditional      *verbTenseGroup   `json:"conditional,omitempty"`
	Imperative       map[string]string `json:"imperative,omitempty"`
}

func main() {
	var (
		inputPath  = flag.String("input", "input.csv", "path to input CSV")
		cacheDir   = flag.String("cache-dir", "cache", "directory for cached JSON and audio files")
		outputPath = flag.String("output", "flashcards.json", "path to generated flashcard JSON")
		refresh    = flag.Bool("refresh", false, "re-download API responses and audio even if cache exists")
	)
	flag.Parse()

	ctx := context.Background()
	client := &http.Client{Timeout: 30 * time.Second}

	baseDir := "."
	if wd, err := os.Getwd(); err == nil {
		baseDir = wd
	}

	jsonCacheDir := filepath.Join(*cacheDir, "json")
	audioCacheDir := filepath.Join(*cacheDir, "audio")
	if err := os.MkdirAll(jsonCacheDir, 0o755); err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(audioCacheDir, 0o755); err != nil {
		fatal(err)
	}

	ids, err := readIDs(*inputPath)
	if err != nil {
		fatal(err)
	}

	cards := make([]flashcard, 0)
	for _, id := range ids {
		resp, jsonPath, audioPath, err := loadOrFetchEntry(ctx, client, id, jsonCacheDir, audioCacheDir, *refresh)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", id, err)
			continue
		}

		entryCards := buildFlashcards(resp, relPath(baseDir, jsonPath), relPath(baseDir, audioPath))
		cards = append(cards, entryCards...)
	}

	outputBytes, err := json.MarshalIndent(cards, "", "  ")
	if err != nil {
		fatal(err)
	}

	if err := os.WriteFile(*outputPath, outputBytes, 0o644); err != nil {
		fatal(err)
	}

	fmt.Printf("processed %d ids -> %d flashcards\n", len(ids), len(cards))
	fmt.Printf("output: %s\n", relPath(baseDir, *outputPath))
	fmt.Printf("json cache: %s\n", relPath(baseDir, jsonCacheDir))
	fmt.Printf("audio cache: %s\n", relPath(baseDir, audioCacheDir))
}

func readIDs(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open csv %s: %w", path, err)
	}
	defer file.Close()

	r := csv.NewReader(file)
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv %s: %w", path, err)
	}
	if len(records) <= 1 {
		return nil, errors.New("input csv has no data rows")
	}

	ids := make([]string, 0, len(records)-1)
	for i, row := range records[1:] {
		if len(row) == 0 {
			continue
		}
		id := strings.TrimSpace(strings.TrimPrefix(row[0], "\uFEFF"))
		if id == "" {
			fmt.Fprintf(os.Stderr, "skip row %d: empty id\n", i+2)
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func loadOrFetchEntry(ctx context.Context, client *http.Client, id, jsonCacheDir, audioCacheDir string, refresh bool) (apiResponse, string, string, error) {
	jsonPath := filepath.Join(jsonCacheDir, id+".json")
	audioPath := ""

	var body []byte
	var err error
	if !refresh {
		body, err = os.ReadFile(jsonPath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return apiResponse{}, "", "", fmt.Errorf("read cache %s: %w", jsonPath, err)
		}
	}

	if len(body) == 0 {
		body, err = fetchEntry(ctx, client, id)
		if err != nil {
			return apiResponse{}, "", "", err
		}
		if err := os.WriteFile(jsonPath, body, 0o644); err != nil {
			return apiResponse{}, "", "", fmt.Errorf("write cache %s: %w", jsonPath, err)
		}
	}

	var resp apiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return apiResponse{}, "", "", fmt.Errorf("decode %s: %w", id, err)
	}

	if resp.Entry.LodID == "" {
		return apiResponse{}, "", "", fmt.Errorf("empty entry in response for %s", id)
	}

	if resp.Entry.AudioFiles.OGG != "" {
		audioPath = filepath.Join(audioCacheDir, id+".ogg")
		if refresh || !fileExists(audioPath) {
			if err := downloadFile(ctx, client, resp.Entry.AudioFiles.OGG, audioPath); err != nil {
				return apiResponse{}, "", "", fmt.Errorf("download audio for %s: %w", id, err)
			}
		}
	}

	return resp, jsonPath, audioPath, nil
}

func fetchEntry(ctx context.Context, client *http.Client, id string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", apiBaseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request for %s: %w", id, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "lod-import/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("request %s returned %s: %s", url, resp.Status, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response %s: %w", url, err)
	}
	return body, nil
}

func downloadFile(ctx context.Context, client *http.Client, url, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build download request %s: %w", url, err)
	}
	req.Header.Set("User-Agent", "lod-import/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("download %s returned %s: %s", url, resp.Status, strings.TrimSpace(string(body)))
	}

	tmpPath := path + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", tmpPath, err)
	}

	var copyErr error
	if _, err := io.Copy(file, resp.Body); err != nil {
		copyErr = fmt.Errorf("write %s: %w", tmpPath, err)
	}
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(tmpPath)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close %s: %w", tmpPath, closeErr)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename %s to %s: %w", tmpPath, path, err)
	}
	return nil
}

func buildFlashcards(resp apiResponse, cachedJSONPath, cachedAudioPath string) []flashcard {
	entry := resp.Entry
	partOfSpeech := normalizePartOfSpeech(entry)

	selectedMeaning, ok := pickMeaningForImport(entry)
	if !ok {
		fmt.Fprintf(os.Stderr, "no meanings found for %s; creating placeholder card\n", entry.LodID)
		return []flashcard{{
			LodID:          entry.LodID,
			NativeLanguage: "TODO",
			TargetLanguage: entry.Lemma,
			PartOfSpeech:   partOfSpeech,
			VerbForms:      buildVerbForms(entry),
			EntryAudioFile: cachedAudioPath,
			CachedResponse: cachedJSONPath,
		}}
	}

	native, clarifier := englishMeaning(selectedMeaning)
	if native == "" {
		native = "TODO"
		fmt.Fprintf(os.Stderr, "missing English translation for %s (meaning %s, number %d)\n", entry.LodID, selectedMeaning.MeaningID, selectedMeaning.Number)
	}

	target := entry.Lemma
	if strings.TrimSpace(selectedMeaning.SecondaryHeadword) != "" {
		target = strings.TrimSpace(selectedMeaning.SecondaryHeadword)
	}

	return []flashcard{{
		LodID:            entry.LodID,
		MeaningID:        selectedMeaning.MeaningID,
		MeaningNumber:    selectedMeaning.Number,
		NativeLanguage:   native,
		TargetLanguage:   target,
		PartOfSpeech:     partOfSpeech,
		ExampleSentences: exampleSentences(selectedMeaning.Examples),
		EnglishClarifier: clarifier,
		VerbForms:        buildVerbForms(entry),
		EntryAudioFile:   cachedAudioPath,
		CachedResponse:   cachedJSONPath,
	}}
}

func buildVerbForms(entry entry) *verbForms {
	if strings.TrimSpace(entry.PartOfSpeech) != "VRB" && strings.TrimSpace(entry.PartOfSpeechLbl) != "VRB" {
		return nil
	}

	forms := &verbForms{
		NRuleForm:       strings.TrimSpace(entry.NRuleForm),
		Infinitive:      strings.TrimSpace(entry.Lemma),
		PastParticiples: firstPastParticiples(entry.MicroStructures),
		AuxiliaryVerbs:  firstAuxiliaryVerbs(entry.MicroStructures),
	}

	if vc := entry.Tables.VerbConjugation; vc != nil {
		if strings.TrimSpace(vc.Infinitive) != "" {
			forms.Infinitive = strings.TrimSpace(vc.Infinitive)
		}
		if len(vc.PastParticiple) > 0 {
			forms.PastParticiples = dedupeStrings(append(forms.PastParticiples, vc.PastParticiple...))
		}
		if len(vc.AuxiliaryVerb) > 0 {
			forms.AuxiliaryVerbs = dedupeStrings(append(forms.AuxiliaryVerbs, vc.AuxiliaryVerb...))
		}
		forms.ConjugationID = strings.TrimSpace(vc.Attributes.ID)
		forms.ConjugationModel = strings.TrimSpace(vc.Attributes.Model)
		forms.SeparableVerb = strings.TrimSpace(vc.Attributes.SeparableVerb)
		forms.Indicative = nonEmptyVerbTenseGroup(vc.Indicative)
		forms.Conditional = nonEmptyVerbTenseGroup(vc.Conditional)
		forms.Imperative = copyMap(vc.Imperative.Present)
	}

	if isEmptyVerbForms(forms) {
		return nil
	}
	return forms
}

func pickMeaningForImport(entry entry) (meaning, bool) {
	var fallback meaning
	foundFallback := false

	for _, ms := range entry.MicroStructures {
		for _, gu := range ms.GrammaticalUnits {
			for _, m := range gu.Meanings {
				if !foundFallback {
					fallback = m
					foundFallback = true
				}
				if m.Number == 1 {
					return m, true
				}
			}
		}
	}

	return fallback, foundFallback
}

func normalizePartOfSpeech(entry entry) string {
	if strings.HasPrefix(entry.PartOfSpeechLbl, "SUBST+") {
		return entry.PartOfSpeechLbl
	}
	if strings.TrimSpace(entry.PartOfSpeech) != "" {
		return entry.PartOfSpeech
	}
	return entry.PartOfSpeechLbl
}

func englishMeaning(meaning meaning) (native, clarifier string) {
	en := meaning.TargetLanguages["en"]
	translations := make([]string, 0)
	clarifiers := make([]string, 0)

	for _, part := range en.Parts {
		content := strings.TrimSpace(part.Content)
		if content == "" {
			continue
		}
		switch part.Type {
		case "translation":
			translations = append(translations, content)
		case "semanticClarifier":
			clarifiers = append(clarifiers, content)
		}
	}

	return strings.Join(translations, "; "), strings.Join(clarifiers, "; ")
}

func exampleSentences(examples []example) []string {
	results := make([]string, 0, len(examples))
	seen := make(map[string]struct{}, len(examples))

	for _, ex := range examples {
		text := renderExample(ex)
		if text == "" {
			continue
		}
		if _, ok := seen[text]; ok {
			continue
		}
		seen[text] = struct{}{}
		results = append(results, text)
	}

	return results
}

func firstPastParticiples(microStructures []microStructure) []string {
	for _, ms := range microStructures {
		if len(ms.PastParticiple) > 0 {
			return dedupeStrings(ms.PastParticiple)
		}
	}
	return nil
}

func firstAuxiliaryVerbs(microStructures []microStructure) []string {
	for _, ms := range microStructures {
		if strings.TrimSpace(ms.AuxiliaryVerb) != "" {
			return []string{strings.TrimSpace(ms.AuxiliaryVerb)}
		}
	}
	return nil
}

func dedupeStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func copyMap(src stringMap) map[string]string {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]string, len(src))
	for k, v := range src {
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			continue
		}
		out[k] = trimmed
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func nonEmptyVerbTenseGroup(group verbTenseGroup) *verbTenseGroup {
	result := &verbTenseGroup{
		Present:        copyMap(group.Present),
		PresentPerfect: copyMap(group.PresentPerfect),
		PastPerfect:    copyMap(group.PastPerfect),
	}
	if len(result.Present) == 0 && len(result.PresentPerfect) == 0 && len(result.PastPerfect) == 0 {
		return nil
	}
	return result
}

func isEmptyVerbForms(forms *verbForms) bool {
	if forms == nil {
		return true
	}
	return forms.NRuleForm == "" &&
		forms.Infinitive == "" &&
		len(forms.PastParticiples) == 0 &&
		len(forms.AuxiliaryVerbs) == 0 &&
		forms.ConjugationID == "" &&
		forms.ConjugationModel == "" &&
		forms.SeparableVerb == "" &&
		forms.Indicative == nil &&
		forms.Conditional == nil &&
		len(forms.Imperative) == 0
}

func renderExample(ex example) string {
	segments := make([]string, 0)
	for _, part := range ex.Parts {
		if part.Type != "text" {
			continue
		}
		segment := renderTokens(part.Parts)
		if segment != "" {
			segments = append(segments, segment)
		}
	}
	return strings.TrimSpace(strings.Join(segments, " "))
}

func renderTokens(tokens []exampleToken) string {
	var b strings.Builder
	first := true
	for _, token := range tokens {
		if token.Type == "attribute" {
			continue
		}
		content := strings.TrimSpace(token.Content)
		if content == "" {
			continue
		}
		if first {
			b.WriteString(content)
			first = false
			continue
		}
		if token.JoinWithPreviousWord {
			b.WriteString(content)
		} else {
			b.WriteByte(' ')
			b.WriteString(content)
		}
	}
	return b.String()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func relPath(base, path string) string {
	if path == "" {
		return ""
	}
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return path
	}
	return rel
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
