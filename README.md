![](https://badgen.net/docker/pulls/pluja/web-whisper-backend)
![](https://badgers.space/badge/License/GPLv3/green)

## Web Whisper Plus

Web Whisper Plus is an audio transcription and subtitling suite that works directly from a web interface in your browser. Seamlessly transcribe, subtitle, translate and edit any media content in any language. Enjoy complete control and privacy with 100% self-hosted, open-source technology.

> ⚠️ The project is in development, so updates may break things until the project is stable. For this, make sure to check the [changelog](https://codeberg.org/pluja/web-whisper-plus/releases) before updating (or if after updating you find any issues).

## Contents

- [Demo video](#demo-video)
- [Changelog & updates](https://codeberg.org/pluja/web-whisper-plus/releases)
- [Features](#features)
    - [Roadmap](#roadmap)
- [Self-hosting](#self-hosting)
    - [GPU support](#gpu-support)
    - [MacOS support](#macos-support)
- [Credits](#credits)
- [Support](#support)

#### Demo video

<details>
    <summary><b>Watch demo video!</b></summary>
<video autoplay controls src="https://codeberg.org/pluja/web-whisper-plus/raw/branch/main/misc/web-whisper-demo.mp4"></video>
</details>

---

## Features

- [x] Transcribe any video or audio file to text
    - From a local file
    - Record from micorphone
    - From a URL source (YouTube, Twitter, Vimeo, etc.) via [yt-dlp](https://github.com/yt-dlp/yt-dlp)
- [x] Generate subtitles for any video or audio file
    - Export in multiple formats: .srt, .txt, .json, copy to clipboard...
    - Edit the generated subtitles with a simple and intuitive interface
- [x] Job queue for batch processing files.
    - SQLite database to save transcription history and data.
- [x] 100% Local processing: no data is sent to any server.
- [x] Auto translation: translate the generated subtitles to any language.

> See version changes in the [Changelog](https://codeberg.org/pluja/web-whisper-plus/src/branch/main/CHANGELOG.md)

### Roadmap

- [ ] Add GPU support
- [ ] Modify words per line in the generated subtitles
- [ ] Auto dubbing: generate a voice over from the generated subtitles in any language.
    - Option to dub locally: using [Piper](https://github.com/rhasspy/piper)
    - Option to use 3rd party ElevenLabs API.
- [ ] OpenAI API - Use the OpenAI API instead of local inference

---

## 🪺 Self-hosting

> You need to have <a href="https://docs.docker.com/engine/install/">docker</a> and <a href="https://docs.docker.com/compose/install/linux/">docker compose</a> installed.

1. Clone this repository: `git clone https://codeberg.org/pluja/web-whisper-plus.git`

2.  `cp example.env .env` - Create a `.env` and edit it to your needs:
    - `WHISPER_MODELS`: Comma separated list of models to download. Available models are `tiny.en`, `tiny`, `small.en`, `small`, `base.en`, `base`, `medium.en`, `medium`, `large-v1`, `large-v2`.
    - `LT_LOAD_ONLY`: List of two letter codes of the languages to load for translation. If you want to load all languages, leave it empty. If you want to load only some languages, separate them with a comma. Example: `en,fr,de`.
        - List of available languages [here](https://libretranslate.com/docs/#/translate/get_languages)
        - If you load all languages, the first time you run the app it will take a while to download all the models. It is recommended to load only the languages you need.
    - `DB_BACKEND`: It defaults to `mysql` which expects the `database` service from the default `docker-compose.yml` to be there. You can also set it to `sqlite`, but it won't work for ARM devices as it needs CGO.
    
3.  Build and start the containers: `docker compose up --build -d`
    - The build process is lightweight, so it should be fast for most devices.
    
> ⚠️ The first run may need a bit of waiting. First, it needs to build the containers. Once these start, the translation and whisper models need to be downloaded. Just wait for around 7 minutes for everything to be ready! (depending on your machine, it could be more)

Now you can visit [http://localhost:8899](http://localhost:8899) and use the app.

- In `db_data` you will find the database files.
- In `whisper_uploads` you will find the uploaded files.

If you want to use a reverse proxy, like Caddy, Nginx or Traefik, check [this issue comment](https://codeberg.org/pluja/web-whisper-plus/issues/1#issuecomment-919230)

### 🔥 GPU support

> Since I reimplemented the whisper-asr backend myself, the GPU support is now on the works, but it will be ready to use very soon!

<!--<details>
    <summary><b>🔥 Using GPU</b></summary>
    <p>First, download the file from this repository.</p>
    <p>Then, just the same steps as above, but in step 2 use this command:</p>
<p><code>docker compose -f docker-compose.yml -f docker-compose.gpu.yml up --build -d</code></p>
<blockquote>
<p>! You may encounter issues, please report them by opening an <a href="https://codeberg.org/pluja/web-whisper-plus/issues/new">issue</a>.</p>
</blockquote>

</details>-->


<details>
    <summary><b>🍏 MacOS / 🍓 Raspberry </b></summary>
    
It should work out of the box just as with any other system, but if you encounter any issues, please report them by opening an [issue](https://codeberg.org/pluja/web-whisper-plus/issues/new). As I don't have an ARM device, I can't test it myself, so I would appreciate if you could test it and report any issues.
</details>


### Update

To update the app, just run `git pull && docker compose pull && docker compose up --build -d` and the containers will update to the latest version.

> You can also run `docker system prune -a` to remove all old and unused images and free up some space.

## Credits

- [Faster-Whisper](https://github.com/guillaumekln/faster-whisper) - This reimplementation of OpenAI's Whisper model is used in the transcription backend API.
- [SvelteKit](https://kit.svelte.dev/) - The frontend of the app is based on SvelteKit.
    - [TailwindCSS](https://tailwindcss.com/) - TailwindCSS is used in the frontend for styling.
    - [DaisyUI](https://daisyui.com/) - DaisyUI is used in the frontend for styling along with TailwindCSS.
- [Golang](https://golang.org/) - The backend of the app is written in Go.
    - [Chi router](https://go-chi.io/#/) - The awesome Chi router is used in the backend.
    - [GORM](https://gorm.io/) - GORM is used to interact with the database.
- [ahmetoner/whisper-asr-webservice](https://github.com/ahmetoner/whisper-asr-webservice) - The backend of the app was once based on this project, so thanks to ahmetoner for his work as it allowed me to start this project. Right now, the whisper api backend was completely rewritten by me and is independent from this project.

---

## Support

If you like my work, you can support me:

- `XMR`: `89vNMC8nQ3WCdCM8As8y7UeNKSqKMTnDdfWKSgV72vyhYYkv7bJS9oYhtTDCWpC5abEVf3MBTkRSrfLbeFpkpwmf2kS7VKu`
- `BTC`: `bc1qesxr2hp3ulqpzx5jwhk04en2p3cqw3kqmsgrdf`


---

## Disclaimer

I am working on this project on my free time and I am not a professional developer, so the code may not be the best. If you find any issues, please report them by opening an [issue](https://codeberg.org/pluja/web-whisper-plus/issues/new) or contribute by opening a [pull request](https://codeberg.org/pluja/web-whisper-plus/pulls).