<script>
	import { timeAgo } from '$lib/utils.js';
    import { onMount, onDestroy } from 'svelte';
    import { jobs } from '$lib/stores';
    import { backendUrl } from '$lib/utils';
    import { toast } from '@zerodevx/svelte-toast'

    let fileInput, urlInput, fetchingJobs = false;

    let selectedModel="small";
    let selectedDevice="cpu";
    let inputSource="file";
    let userPrompt = "";

    const languages = ["auto", "af", "am", "ar", "as", "az", "ba", "be", "bg", "bn", "bo", "br", "bs", "ca", "cs", "cy", "da", "de", "el", "en", "es", "et", "eu", "fa", "fi", "fo", "fr", "gl", "gu", "ha", "haw", "he", "hi", "hr", "ht", "hu", "hy", "id", "is", "it", "ja", "jw", "ka", "kk", "km", "kn", "ko", "la", "lb", "ln", "lo", "lt", "lv", "mg", "mi", "mk", "ml", "mn", "mr", "ms", "mt", "my", "ne", "nl", "nn", "no", "oc", "pa", "pl", "ps", "pt", "ro", "ru", "sa", "sd", "si", "sk", "sl", "sn", "so", "sq", "sr", "su", "sv", "sw", "ta", "te", "tg", "th", "tk", "tl", "tr", "tt", "uk", "ur", "uz", "vi", "yi", "yo", "zh"]
    let selectedLanguage="auto";
    let translationLanguage = '';
    let selectedLanguages = Array($jobs.length).fill(''); // Array of selected languages for translation

    let isRecording = false;
    let audioRecorded = false;
    let audioFilename = "";
    let mediaRecorder;
    let recordedChunks = [];

    // Jobs pagination
    let currentPage = 1;
    const jobsPerPage = 4;

    const totalPages = () => Math.ceil($jobs.length / jobsPerPage);
    function goToPage(page) { if (page >= 1 && page <= totalPages()) currentPage = page; }

    async function startRecording() {
        recordedChunks = [];
        isRecording = true;
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true, video: false });
        mediaRecorder = new MediaRecorder(stream);
        mediaRecorder.ondataavailable = e => recordedChunks.push(e.data);
        mediaRecorder.start();
    }

    function stopRecording() {
        mediaRecorder.stop();
        isRecording = false;
        audioRecorded = true;
    }


    const copyToClipboard = (job, language) => {
        if(language == "") navigator.clipboard.writeText(job.result.text);
        job.translations.forEach(item => {
            if(item.target_language == language) {
                navigator.clipboard.writeText(item.result.text);
            }
        });
        toast.push('Text copied to clipboard!')
    }

    const handleTranslate = id => {
        if(translationLanguage) {
            const url = `${backendUrl}/translate/${id}/${translationLanguage}`;
            fetch(url)
            .then(() => toast.push('Translation started!'))
            .catch(error => {
                console.error(error);
                toast.push('Error translating text!', { classes: ['bg-error'] })
            });
        }
    }

    async function sendForm() {
        let formData = new FormData();

        
        if (inputSource === "record") {
            formData.append("source", "file");
            let filename = audioFilename == "" ? "recording" : audioFilename.replace(/\s+/g, "_");
            formData.append("file", new Blob(recordedChunks), `${filename}.wav`);
        } else if (inputSource === "file") {
            formData.append("source", inputSource);
            formData.append("file", fileInput.files[0]);
        } else if (inputSource === "url") {
            formData.append("source", inputSource);
            formData.append("url", urlInput);
        }
        formData.append("language", selectedLanguage);
        formData.append("model_size", selectedModel);
        formData.append("device", selectedDevice);
        formData.append("task", "transcribe");
        
        fetch(`${backendUrl}/asr`, {
            method: "POST",
            body: formData
        });

        toast.push('New job created!')
        // wait 200ms
        await new Promise(r => setTimeout(r, 200));
        await fetchJobs();
    }
    
    async function removeJob(id) {
        const updatedJobs = $jobs.filter(job => job.ID !== id);
        jobs.set(updatedJobs);
        fetch(`${backendUrl}/jobs/delete/${id}`, {method: "GET"});
    }

    const fetchJobs = async () => {
        if (fetchJobsInProgress) return;
        fetchJobsInProgress = true;
        const res = await fetch(`${backendUrl}/jobs`);
        const data = await res.json();
        if (data) {
            jobs.set(data);
        }
        fetchJobsInProgress = false;
    };

    function getAvailableTargetLanguages(lang) {
        const langObj = availableLanguages.find(e => e.code === lang);
        return langObj ? langObj.targets : [];
    }


    let fetchJobsInProgress = false;
    let availableLanguages = [];
    const getAvailableLangs = () => {
        const fetchLanguages = () => {
            fetch(`${backendUrl}/languages`)
            .then(res => res.json())
            .then(data => {
                if (data) {
                    availableLanguages = data;
                    // Languages fetched successfully, stop trying
                    clearInterval(fetchLanguagesInterval);
                }
            });
        };

        // Fetch languages repeatedly until successful
        const fetchLanguagesInterval = setInterval(fetchLanguages, 5000);
        fetchLanguages();
    };


    let loading = false;
    onMount(async () => {
        loading = true;
        await fetchJobs();
        setInterval(fetchJobs, 5000);
        loading = false;
        await getAvailableLangs();
    });

    $: currentJobs = $jobs.slice((currentPage - 1) * jobsPerPage, currentPage * jobsPerPage);
</script>

<header class="text-center py-4">
    <h1 class="font-bold text-3xl">Web Whisper+</h1>
    <h2 class="text-lg">AI transcription suite</h2>
</header>

{#if loading}
    <div class="text-center py-32">
        <h1 class="font-bold text-2xl">Loading...</h1>
    </div>
{:else}
<div class="flex flex-col items-center justify-center">
    {#if availableLanguages.length == 0}
        <div class="alert alert-warning max-w-md">
            <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
            <span>Waiting translation backend... Translation not available yet.</span>
        </div>
    {/if}
    <div class="grid grid-cols-1 md:grid-cols-2 gap-8 lg:gap-16 w-full max-w-7xl mx-auto">
        <!--FORM-->
        <div class="w-full p-4">
            <section class="flex w-full flex-col items-center my-12">
                <!--Source Chooser-->
                <div class="flex">
                    <div class="form-control">
                        <label class="label cursor-pointer flex flex-col">
                          <span class="label-text">File</span> 
                          <input bind:group={inputSource} value="file" type="radio" name="radio-10" class="radio checked:bg-red-500" />
                        </label>
                    </div>
                    <div class="form-control">
                        <label class="label cursor-pointer flex flex-col">
                          <span class="label-text">Record</span> 
                          <input bind:group={inputSource} value="record" type="radio" name="radio-10" class="radio checked:bg-blue-500" />
                        </label>
                    </div>
                    <div class="form-control">
                        <label class="label cursor-pointer flex flex-col">
                          <span class="label-text">URL</span> 
                          <input bind:group={inputSource} value="url" type="radio" name="radio-10" class="radio checked:bg-blue-500" />
                        </label>
                    </div>
                </div>

                <!--File Source-->
                {#if inputSource == "file"}
                    <div class="max-w-lg text-center">
                        <h3 class="font-bold text-xl my-2">Choose a file to transcribe</h3>
                        <input bind:this={fileInput} type="file" class="file-input file-input-bordered file-input-primary w-full" />
                    </div>
                {/if}

                {#if inputSource == "url"}
                    <div class="max-w-lg text-center py-2">
                        <label class="label">
                            <span class="label-text">yt-dlp compatible source URL</span>
                        </label>
                        <input type="text" bind:value={urlInput} placeholder="https://youtube.com/watch?..." class="input input-bordered w-full max-w-xs" />
                    </div>
                {/if}

                <!--Record Source-->
                {#if inputSource == "record"}
                    <div class="max-w-lg text-center">
                        <h3 class="font-bold text-xl my-2">Record an audio</h3>
                        <div class="flex flex-col justify-center items-center">
                        {#if !isRecording}
                            <div class="form-control w-full max-w-xs mb-2">
                                <label class="label">
                                <span class="label-text">Recording name</span>
                                </label>
                                <input type="text" bind:value={audioFilename} placeholder="Type here" class="input input-xs input-bordered w-full max-w-xs" />
                            </div>
                            <button class="btn btn-info" on:click={startRecording}>
                                {#if audioRecorded}
                                    Record Again
                                {:else}
                                    Record
                                {/if}
                            </button>
                        {:else}
                            <button class="btn btn-error" on:click={stopRecording}>Stop</button>
                        {/if}
                        </div>
                    </div>
                {/if}
            
                <div class="max-w-lg w-full text-center mt-8">
                    <h3 class="font-bold text-xl my-2">Configure whisper settings</h3>
            
                    <div class="form-control w-full">
                        <label class="label">
                          <span class="label-text">Model</span>
                        </label>
                        <select bind:value={selectedModel} class="select select-bordered w-full">
                            <option selected value="small">Small</option>
                            <option value="tiny">Tiny</option>
                            <option value="base">Base</option>
                            <option value="medium">Medium</option>
                            <option value="large-v2">Large-v2</option>
                        </select>
                    </div>
            
                    <div class="form-control w-full">
                        <label class="label">
                          <span class="label-text">Language</span>
                        </label>
                        <select bind:value={selectedLanguage} class="select select-bordered w-full">
                            <option value="" disabled selected>Select a language</option>
                            {#each languages as language}
                                <option value="{language}">{language}</option>
                            {/each}
                        </select>
                    </div>

                    <div class="form-control w-full">
                        <label class="label">
                          <span class="label-text">Device</span>
                        </label>
                        <select bind:value={selectedDevice} class="select select-bordered w-full">
                            <option value="cpu">CPU</option>
                            <option disabled value="cuda">CUDA (not available yet)</option>
                        </select>
                    </div>
            
                    <button on:click={sendForm} class="btn mt-4 btn-info bg-opacity-80 hover:bg-opacity-100">Transcribe</button>
                </div>
            </section>
        </div>
        
        <!--JOBS-->
        <div class="w-full max-w-lg p-4">
            <section class="flex w-full flex-col justify-center items-center">
                {#if $jobs.length == 0}
                    <h2 class="text-2xl font-bold text-center mb-4 mt-36">No jobs yet</h2>
                {/if}
                {#if $jobs.length > 0}
                    <h2 class="text-2xl font-bold text-center mb-4">Jobs</h2>
                    <div class="w-full">
                        <table class="table">
                        <!-- head -->
                        <thead class="sticky head">
                            <tr>
                            <th></th>
                            <th>Status</th>
                            <th>Filename</th>
                            <th>Language</th>
                            <th>Date</th>
                            <th></th>
                            </tr>
                        </thead>
                        <tbody class="overflow-y-auto max-h-64">
                            {#each currentJobs as job, index (job.ID)}
                                <tr class="text-center text-md">
                                    <th>{job.ID}</th>
                                    <td> 
                                        {#if job.job_status == "0"}
                                            <span class="loading loading-spinner loading-md"></span>
                                        {:else if job.job_status == "1"}
                                            <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-success" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                                                <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                                                <path d="M17 3.34a10 10 0 1 1 -14.995 8.984l-.005 -.324l.005 -.324a10 10 0 0 1 14.995 -8.336zm-1.293 5.953a1 1 0 0 0 -1.32 -.083l-.094 .083l-3.293 3.292l-1.293 -1.292l-.094 -.083a1 1 0 0 0 -1.403 1.403l.083 .094l2 2l.094 .083a1 1 0 0 0 1.226 0l.094 -.083l4 -4l.083 -.094a1 1 0 0 0 -.083 -1.32z" stroke-width="0" fill="currentColor"></path>
                                            </svg>
                                        {:else if job.job_status == "2"}
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-error" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                                                <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                                                <path d="M17 3.34a10 10 0 1 1 -14.995 8.984l-.005 -.324l.005 -.324a10 10 0 0 1 14.995 -8.336zm-6.489 5.8a1 1 0 0 0 -1.218 1.567l1.292 1.293l-1.292 1.293l-.083 .094a1 1 0 0 0 1.497 1.32l1.293 -1.292l1.293 1.292l.094 .083a1 1 0 0 0 1.32 -1.497l-1.292 -1.293l1.292 -1.293l.083 -.094a1 1 0 0 0 -1.497 -1.32l-1.293 1.292l-1.293 -1.292l-.094 -.083z" stroke-width="0" fill="currentColor"></path>
                                            </svg>
                                        {/if}
                                    </td>
                                    <td class="break-words">{job.file_name.split('_').slice(0, 7).join(' ')}</td>
                                    <td>
                                        <select bind:value={selectedLanguages[index]} class="select" name="" id="">
                                            <option selected value="">✅ {job.language}</option>
                                            {#if availableLanguages.length > 0}
                                                {#each job.translations as translation}
                                                    <option value="{translation.target_language}">🌐 {translation.target_language}</option>
                                                {/each}
                                            {/if}
                                        </select>
                                    </td>
                                    <td>{timeAgo(job.UpdatedAt)}</td>
                                    <td class="flex flex-wrap items-center justify-center align-middle space-x-1 space-y-1">
                                        {#if job.job_status == "1"}
                                            <!-- ACTION DOWNLOAD SUBS -->
                                            <div class="dropdown dropdown-left tooltip" data-tip="Download">
                                                <label for="downloadSubs" aria-label="Export" tabindex="0" class="btn btn-info btn-xs m-1 flex justify-center align-middle">
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                                                        <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                                                        <path d="M4 17v2a2 2 0 0 0 2 2h12a2 2 0 0 0 2 -2v-2"></path>
                                                        <path d="M7 11l5 5l5 -5"></path>
                                                        <path d="M12 4l0 12"></path>
                                                    </svg>
                                                </label>
                                                <ul tabindex="0" class="dropdown-content menu p-2 shadow bg-base-300 text-neutral-300 rounded-box w-48">
                                                    <li><a href="{backendUrl}/download/{job.ID}/srt?lang={selectedLanguages[index]}">As .srt</a></li>
                                                    <li><a href="{backendUrl}/download/{job.ID}/txt?lang={selectedLanguages[index]}">As .txt</a></li>
                                                    <li><a href="{backendUrl}/download/{job.ID}/json?lang={selectedLanguages[index]}">As .json</a></li>
                                                    <li><span on:keyup={copyToClipboard(job, selectedLanguages[index])} on:click={copyToClipboard(job, selectedLanguages[index])}>Copy raw text</span></li>                                                    
                                                </ul>
                                            </div>
                                            

                                            <!--TODO: Implement Autodubbing-->
                                            <!-- ACTION AUTODUB SUBS -->
                                            <span class="btn btn-disabled hidden btn-xs btn-primary">Dub</span>

                                            <!-- ACTION EDIT SUBS -->
                                            <a href="/edit/{job.ID}/{selectedLanguages[index]}" class="btn btn-xs btn-primary tooltip" data-tip="Edit subtitles">
                                                <span aria-label="Edit">
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                                                        <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                                                        <path d="M7 7h-1a2 2 0 0 0 -2 2v9a2 2 0 0 0 2 2h9a2 2 0 0 0 2 -2v-1"></path>
                                                        <path d="M20.385 6.585a2.1 2.1 0 0 0 -2.97 -2.97l-8.415 8.385v3h3l8.385 -8.415z"></path>
                                                        <path d="M16 5l3 3"></path>
                                                     </svg>
                                                </span>
                                            </a>

                                            <!-- ACTION TRANSLATE SUBS -->
                                            {#if availableLanguages.length > 0}
                                                <div class="dropdown dropdown-left">
                                                    <label for="translateSubs" tabindex="-1" class="btn btn-primary btn-xs">
                                                        <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                                                            <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                                                            <path d="M4 5h7"></path>
                                                            <path d="M7 4c0 4.846 0 7 .5 8"></path>
                                                            <path d="M10 8.5c0 2.286 -2 4.5 -3.5 4.5s-2.5 -1.135 -2.5 -2c0 -2 1 -3 3 -3s5 .57 5 2.857c0 1.524 -.667 2.571 -2 3.143"></path>
                                                            <path d="M12 20l4 -9l4 9"></path>
                                                            <path d="M19.1 18h-6.2"></path>
                                                        </svg>
                                                    </label>
                                                    <div tabindex="-1" class="dropdown-content card card-compact p-2 shadow bg-base-300 text-primary-content">
                                                        <div class="card-body">
                                                            <h3 class="text-lg font-bold">Translate {job.file_name}</h3>
                                                            <div class="form-control w-full max-w-xs">
                                                                <label for="pickSubsTranslationLanguage" class="label">
                                                                <span class="label-text">Pick the target language</span>
                                                                </label>
                                                                <select bind:value={translationLanguage} class="select select-bordered">
                                                                <option disabled selected value="">Pick language</option>
                                                                    {#each getAvailableTargetLanguages(job.language) as lang}
                                                                        <option value="{lang}">{lang}</option>
                                                                    {/each}
                                                                </select>
                                                            </div>
                                                            <button disabled={translationLanguage == ''} on:click={handleTranslate(job.ID)} class="btn btn-primary btn-wide">Translate</button>
                                                        </div>
                                                    </div>
                                                </div>
                                            {/if}
                                            
                                            <!-- ACTION REMOVE SUBS -->
                                            <span aria-label="Remove" on:keyup={() => removeJob(job.ID)} on:click={() => removeJob(job.ID)} class="btn btn-xs btn-error font-bold text-black tooltip" data-tip="Remove job">
                                                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-red-900" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                                                    <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                                                    <path d="M20 6a1 1 0 0 1 .117 1.993l-.117 .007h-.081l-.919 11a3 3 0 0 1 -2.824 2.995l-.176 .005h-8c-1.598 0 -2.904 -1.249 -2.992 -2.75l-.005 -.167l-.923 -11.083h-.08a1 1 0 0 1 -.117 -1.993l.117 -.007h16zm-9.489 5.14a1 1 0 0 0 -1.218 1.567l1.292 1.293l-1.292 1.293l-.083 .094a1 1 0 0 0 1.497 1.32l1.293 -1.292l1.293 1.292l.094 .083a1 1 0 0 0 1.32 -1.497l-1.292 -1.293l1.292 -1.293l.083 -.094a1 1 0 0 0 -1.497 -1.32l-1.293 1.292l-1.293 -1.292l-.094 -.083z" stroke-width="0" fill="currentColor"></path>
                                                    <path d="M14 2a2 2 0 0 1 2 2a1 1 0 0 1 -1.993 .117l-.007 -.117h-4l-.007 .117a1 1 0 0 1 -1.993 -.117a2 2 0 0 1 1.85 -1.995l.15 -.005h4z" stroke-width="0" fill="currentColor"></path>
                                                </svg>
                                            </span>
                                        {:else if job.job_status == "2"}
                                            <!-- JOB HAS FAILED -->
                                            <span class="badge badge-error">Failed</span>
                                            <span on:keyup={() => removeJob(job.ID)} on:click={() => removeJob(job.ID)} class="btn btn-xs btn-error font-bold text-black tooltip" data-tip="Remove job">
                                                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-red-900" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                                                    <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                                                    <path d="M20 6a1 1 0 0 1 .117 1.993l-.117 .007h-.081l-.919 11a3 3 0 0 1 -2.824 2.995l-.176 .005h-8c-1.598 0 -2.904 -1.249 -2.992 -2.75l-.005 -.167l-.923 -11.083h-.08a1 1 0 0 1 -.117 -1.993l.117 -.007h16zm-9.489 5.14a1 1 0 0 0 -1.218 1.567l1.292 1.293l-1.292 1.293l-.083 .094a1 1 0 0 0 1.497 1.32l1.293 -1.292l1.293 1.292l.094 .083a1 1 0 0 0 1.32 -1.497l-1.292 -1.293l1.292 -1.293l.083 -.094a1 1 0 0 0 -1.497 -1.32l-1.293 1.292l-1.293 -1.292l-.094 -.083z" stroke-width="0" fill="currentColor"></path>
                                                    <path d="M14 2a2 2 0 0 1 2 2a1 1 0 0 1 -1.993 .117l-.007 -.117h-4l-.007 .117a1 1 0 0 1 -1.993 -.117a2 2 0 0 1 1.85 -1.995l.15 -.005h4z" stroke-width="0" fill="currentColor"></path>
                                                </svg>
                                            </span>
                                        {:else}
                                            <!-- WAITING FOR JOB -->
                                            <span class="badge badge-info">Waiting...</span>
                                        {/if}
                                    </td>
                                </tr>
                            {/each}
                        </tbody>
                        </table>
                        {#if $jobs.length > 0}
                            <!-- JOB PAGINATION -->
                            <div class="btn-group">
                                <button on:click={() => goToPage(currentPage - 1)} class="btn">«</button>
                                <button class="btn">Page {currentPage}/{totalPages()}</button>
                                <button on:click={() => goToPage(currentPage + 1)} class="btn">»</button>
                            </div>
                        {/if}
                    </div>
                {/if}
            </section>
        </div>
    </div>
</div>
{/if}