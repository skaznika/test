## 2023.08.01

> Make sure to read the self-hosting instructions in the README.md file, you will need to clone this repo and build the docker image yourself for this release.

### Important Changes

- The app now uses the new whisper-api API backend that was fully written by me. This means that the app is now completely independent from the whisper-asr-webservice project by [ahmetoner](https://github.com/ahmetoner) . The new API backend is much simpler consisting in a few lines of Python code. It makes use of [faster-whisper](https://github.com/guillaumekln/faster-whisper) project.

- New versioning scheme. This is the first release of the new versioning scheme. The version number is now the date of the release. The date is in the format `YYYY.MM.DD`. The releases page will no longer be used, instead, the releases will be listed in the CHANGELOG.md file.

- I will no longer be updating the dockerhub image, as it was bringing many problems with different architectures. I may reconsider pushing the Linux/amd64 image to dockerhub. Instead, you will need to build the image yourself.

### Added

- You can now choose the whisper model to use from the web ui per each execution. This means that you can run a transcription job with the `large-v2` model and another with the `small.en` model!
    - The available models are `tiny.en`, `tiny`, `small.en`, `small`, `base.en`, `base`, `medium.en`, `medium`, `large-v1`, `large-v2`. The default model is `small`.

- Autodetect language: you can either choose the language to force it or let whisper guess the language by selecting the new `auto` option.

- New `whisper-api` returns **word-level** accuracy and **word-level** timings. For now, this is not yet used. But it opens the door to, for example, being able to choose how many words per line you want to have in the transcription subtitle file. This feature will be added in the future. This is already available in the JSON output :)

### Removed

- GPU support is no longer available. I will reimplement it in the near future, but for now, it is not available.