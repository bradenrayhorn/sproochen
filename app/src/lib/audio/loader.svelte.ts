class AudioPlayer {
  constructor() {}

  playAudio = async (id: string) => {
    const audioBuffer = await (await fetch(`/audio/${id}.opus`)).arrayBuffer();

    const ctx = new AudioContext();
    const decoded = await ctx.decodeAudioData(audioBuffer);
    const source = ctx.createBufferSource();
    source.buffer = decoded;
    source.connect(ctx.destination);

    await new Promise<void>((resolve) => {
      source.onended = () => {
        source.disconnect();
        ctx.close();
        resolve();
      };

      source.start();
    });
  };
}

export { AudioPlayer };
