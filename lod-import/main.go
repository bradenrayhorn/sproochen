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
	PartOfSpeech     string            `json:"partOfSpeech"`
	PartOfSpeechLbl  string            `json:"partOfSpeechLabel"`
	AuxiliaryVerb    string            `json:"auxiliaryVerb"`
	PastParticiple   []string          `json:"pastParticiple"`
	Inflection       *inflection       `json:"inflection,omitempty"`
	InternalLinks    []internalLink    `json:"internalLinks"`
	GrammaticalUnits []grammaticalUnit `json:"grammaticalUnits"`
}

type internalLink struct {
	Tag      string `json:"tag"`
	Content  string `json:"content"`
	IDRef    string `json:"idRef"`
	Relation string `json:"relation,omitempty"`
}

type inflection struct {
	Forms []inflectionForm `json:"forms"`
}

type inflectionForm struct {
	Content string `json:"content"`
}

type grammaticalUnit struct {
	Meanings []meaning `json:"meanings"`
}

type meaning struct {
	MeaningID         string                  `json:"meaningID"`
	Number            int                     `json:"number"`
	SecondaryHeadword string                  `json:"secondaryHeadword"`
	Inflection        *inflection             `json:"inflection,omitempty"`
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
	Plural           []string   `json:"plural,omitempty"`
	ExampleSentences []string   `json:"example_sentences,omitempty"`
	EnglishClarifier string     `json:"english_clarifier,omitempty"`
	VerbForms        *verbForms `json:"verb_forms,omitempty"`
	EntryAudioFile   string     `json:"entry_audio_file,omitempty"`
	CachedResponse   string     `json:"cached_response,omitempty"`
}

type flashcardOverrideIndex struct {
	exact   map[string]flashcard
	lodOnly map[string]flashcard
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
		inputPath     = flag.String("input", "input.csv", "path to input CSV")
		cacheDir      = flag.String("cache-dir", "cache", "directory for cached JSON and audio files")
		outputPath    = flag.String("output", "flashcards.json", "path to generated flashcard JSON")
		overridesPath = flag.String("overrides", "overrides.json", "path to optional flashcard overrides JSON")
		refresh       = flag.Bool("refresh", false, "re-download API responses and audio even if cache exists")
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

	overrides, err := loadFlashcardOverrides(*overridesPath)
	if err != nil {
		fatal(err)
	}
	overrideIndex := buildFlashcardOverrideIndex(overrides)

	cards := make([]flashcard, 0)
	for _, id := range ids {
		resp, jsonPath, audioPath, err := loadOrFetchEntry(ctx, client, id, jsonCacheDir, audioCacheDir, *refresh)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", id, err)
			continue
		}

		entryCards := buildFlashcards(resp, relPath(baseDir, jsonPath), relPath(baseDir, audioPath), overrideIndex)
		cards = append(cards, entryCards...)
	}

	if len(overrides) > 0 {
		cards = applyFlashcardOverrides(cards, overrides)
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

func buildFlashcards(resp apiResponse, cachedJSONPath, cachedAudioPath string, overrides flashcardOverrideIndex) []flashcard {
	entry := resp.Entry
	partOfSpeech := normalizePartOfSpeech(entry)
	plural := buildPluralForms(entry)

	selectedMeaning, ok := pickMeaningForImport(entry)
	if !ok {
		// fallback to internal links for a definition
		native := internalLinksTranslation(entry)
		if native == "" {
			if !overrides.matchesLodOnly(entry.LodID) {
				fmt.Fprintf(os.Stderr, "no meanings found for %s; creating placeholder card\n", entry.LodID)
			}
			native = "TODO"
		}

		return []flashcard{{
			LodID:          entry.LodID,
			NativeLanguage: native,
			TargetLanguage: entry.Lemma,
			PartOfSpeech:   partOfSpeech,
			Plural:         plural,
			VerbForms:      buildVerbForms(entry),
			EntryAudioFile: cachedAudioPath,
			CachedResponse: cachedJSONPath,
		}}
	}

	native, clarifier := englishMeaning(selectedMeaning)
	if native == "" {
		native = clarifier
	}
	if native == "" {
		native = "TODO"
		if !overrides.matchesExact(entry.LodID, selectedMeaning.MeaningID) {
			fmt.Fprintf(os.Stderr, "missing English translation for %s (meaning %s, number %d)\n", entry.LodID, selectedMeaning.MeaningID, selectedMeaning.Number)
		}
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
		Plural:           plural,
		ExampleSentences: exampleSentences(selectedMeaning.Examples),
		EnglishClarifier: clarifier,
		VerbForms:        buildVerbForms(entry),
		EntryAudioFile:   cachedAudioPath,
		CachedResponse:   cachedJSONPath,
	}}
}

func loadFlashcardOverrides(path string) ([]flashcard, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil
	}
	body, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read overrides %s: %w", path, err)
	}

	var overrides []flashcard
	if err := json.Unmarshal(body, &overrides); err != nil {
		return nil, fmt.Errorf("decode overrides %s: %w", path, err)
	}
	return overrides, nil
}

func applyFlashcardOverrides(cards []flashcard, overrides []flashcard) []flashcard {
	index := buildFlashcardOverrideIndex(overrides)
	if len(index.exact) == 0 && len(index.lodOnly) == 0 {
		return cards
	}

	for i := range cards {
		if override, ok := index.match(cards[i]); ok {
			mergeFlashcard(&cards[i], override)
		}
	}
	return cards
}

func buildFlashcardOverrideIndex(overrides []flashcard) flashcardOverrideIndex {
	index := flashcardOverrideIndex{
		exact:   make(map[string]flashcard),
		lodOnly: make(map[string]flashcard),
	}
	for _, override := range overrides {
		lodID := strings.TrimSpace(override.LodID)
		if lodID == "" {
			continue
		}
		if strings.TrimSpace(override.MeaningID) == "" {
			index.lodOnly[lodID] = override
			continue
		}
		index.exact[flashcardOverrideKey(override.LodID, override.MeaningID)] = override
	}
	return index
}

func (i flashcardOverrideIndex) matchesExact(lodID, meaningID string) bool {
	_, ok := i.exact[flashcardOverrideKey(lodID, meaningID)]
	return ok
}

func (i flashcardOverrideIndex) matchesLodOnly(lodID string) bool {
	_, ok := i.lodOnly[strings.TrimSpace(lodID)]
	return ok
}

func (i flashcardOverrideIndex) match(card flashcard) (flashcard, bool) {
	if strings.TrimSpace(card.LodID) == "" {
		return flashcard{}, false
	}
	if strings.TrimSpace(card.MeaningID) != "" {
		if override, ok := i.exact[flashcardOverrideKey(card.LodID, card.MeaningID)]; ok {
			return override, true
		}
	}
	if override, ok := i.lodOnly[strings.TrimSpace(card.LodID)]; ok {
		return override, true
	}
	return flashcard{}, false
}

func flashcardOverrideKey(lodID, meaningID string) string {
	return strings.TrimSpace(lodID) + "\x00" + strings.TrimSpace(meaningID)
}

func mergeFlashcard(dst *flashcard, src flashcard) {
	if strings.TrimSpace(src.LodID) != "" {
		dst.LodID = strings.TrimSpace(src.LodID)
	}
	if strings.TrimSpace(src.MeaningID) != "" {
		dst.MeaningID = strings.TrimSpace(src.MeaningID)
	}
	if src.MeaningNumber != 0 {
		dst.MeaningNumber = src.MeaningNumber
	}
	if strings.TrimSpace(src.NativeLanguage) != "" {
		dst.NativeLanguage = src.NativeLanguage
	}
	if strings.TrimSpace(src.TargetLanguage) != "" {
		dst.TargetLanguage = src.TargetLanguage
	}
	if strings.TrimSpace(src.PartOfSpeech) != "" {
		dst.PartOfSpeech = src.PartOfSpeech
	}
	if src.Plural != nil {
		dst.Plural = append([]string(nil), src.Plural...)
	}
	if src.ExampleSentences != nil {
		dst.ExampleSentences = append([]string(nil), src.ExampleSentences...)
	}
	if strings.TrimSpace(src.EnglishClarifier) != "" {
		dst.EnglishClarifier = src.EnglishClarifier
	}
	if src.VerbForms != nil {
		if dst.VerbForms == nil {
			dst.VerbForms = &verbForms{}
		}
		mergeVerbForms(dst.VerbForms, *src.VerbForms)
	}
	if strings.TrimSpace(src.EntryAudioFile) != "" {
		dst.EntryAudioFile = src.EntryAudioFile
	}
	if strings.TrimSpace(src.CachedResponse) != "" {
		dst.CachedResponse = src.CachedResponse
	}
}

func mergeVerbForms(dst *verbForms, src verbForms) {
	if strings.TrimSpace(src.NRuleForm) != "" {
		dst.NRuleForm = src.NRuleForm
	}
	if strings.TrimSpace(src.Infinitive) != "" {
		dst.Infinitive = src.Infinitive
	}
	if src.PastParticiples != nil {
		dst.PastParticiples = append([]string(nil), src.PastParticiples...)
	}
	if src.AuxiliaryVerbs != nil {
		dst.AuxiliaryVerbs = append([]string(nil), src.AuxiliaryVerbs...)
	}
	if strings.TrimSpace(src.ConjugationID) != "" {
		dst.ConjugationID = src.ConjugationID
	}
	if strings.TrimSpace(src.ConjugationModel) != "" {
		dst.ConjugationModel = src.ConjugationModel
	}
	if strings.TrimSpace(src.SeparableVerb) != "" {
		dst.SeparableVerb = src.SeparableVerb
	}
	if src.Indicative != nil {
		if dst.Indicative == nil {
			dst.Indicative = &verbTenseGroup{}
		}
		mergeVerbTenseGroup(dst.Indicative, *src.Indicative)
	}
	if src.Conditional != nil {
		if dst.Conditional == nil {
			dst.Conditional = &verbTenseGroup{}
		}
		mergeVerbTenseGroup(dst.Conditional, *src.Conditional)
	}
	if src.Imperative != nil {
		dst.Imperative = copyStringMap(src.Imperative)
	}
}

func mergeVerbTenseGroup(dst *verbTenseGroup, src verbTenseGroup) {
	if src.Present != nil {
		dst.Present = copyMap(src.Present)
	}
	if src.PresentPerfect != nil {
		dst.PresentPerfect = copyMap(src.PresentPerfect)
	}
	if src.PastPerfect != nil {
		dst.PastPerfect = copyMap(src.PastPerfect)
	}
}

func copyStringMap(src map[string]string) map[string]string {
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

func buildPluralForms(entry entry) []string {
	if strings.TrimSpace(entry.PartOfSpeech) != "SUBST" {
		return nil
	}
	for _, ms := range entry.MicroStructures {
		if partOfSpeech := strings.TrimSpace(ms.PartOfSpeech); partOfSpeech != "" && partOfSpeech != "SUBST" {
			continue
		}
		if forms := pluralFormsFromInflection(ms.Inflection); len(forms) > 0 {
			return forms
		}
		for _, gu := range ms.GrammaticalUnits {
			for _, m := range gu.Meanings {
				if forms := pluralFormsFromInflection(m.Inflection); len(forms) > 0 {
					return forms
				}
			}
		}
	}
	return nil
}

func pluralFormsFromInflection(inf *inflection) []string {
	if inf == nil || len(inf.Forms) == 0 {
		return nil
	}
	forms := make([]string, 0, len(inf.Forms))
	for _, form := range inf.Forms {
		content := strings.TrimSpace(form.Content)
		if content == "" {
			continue
		}
		forms = append(forms, content)
	}
	return dedupeStrings(forms)
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
		case "function":
			clarifiers = append(clarifiers, fmt.Sprintf("{%s}", content))
		}
	}

	return strings.Join(translations, "; "), strings.Join(clarifiers, "; ")
}

func internalLinksTranslation(entry entry) string {
	parts := make([]string, 0)
	for _, ms := range entry.MicroStructures {
		for _, link := range ms.InternalLinks {
			content := strings.TrimSpace(link.Content)
			if content == "" {
				continue
			}
			parts = append(parts, content)
		}
		if len(parts) > 0 {
			break
		}
	}
	return strings.Join(parts, " ")
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
