<script>
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { backendUrl } from '$lib/utils';
	import { toast } from '@zerodevx/svelte-toast';

	let loading = true, videoSrc='', video, jobObj, result;

	onMount(async () => {
		loading = true;
		const res = await fetch(`${backendUrl}/jobs/${$page.params.id}`);
		jobObj = await res.json();
		if ($page.params.lang) {
			jobObj.translations.forEach(e => {
				if (e.target_language == $page.params.lang) result = e.result;
			});
		} else {
			result = jobObj.result;
		}
		loading = false;
	});

	function save() {
		const updateItem = item => {
			item.result = result;
			item.result.text = result.segments.map(s => s[2]).join(' ');
			return item;
		};

		jobObj = $page.params.lang
			? { ...jobObj, translations: jobObj.translations.map(e => e.target_language == $page.params.lang ? updateItem(e) : e) }
			: updateItem(jobObj);

		fetch(`${backendUrl}/jobs/edit`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(jobObj)
		});
		toast.push('Edited!');
	}

	function seek(start) {
		video.currentTime = start;
		video.play();
	}

	$: if (jobObj) videoSrc = `${backendUrl}/file/${jobObj.ID}`;
</script>

{#if loading}
	<h1>Loading...</h1>
{:else if !jobObj}
	<h1>Job not found</h1>
{:else}
	<h1 class="font-bold text-xl pt-4 text-center">{jobObj.file_name}</h1>
	<h2 class="font-bold text-xl pb-4 text-center">
		{#if $page.params.lang}
			<div class="flex flex-row items-center text-yellow-500 justify-center">
				<span>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						class="w-6 h-6"
						width="24"
						height="24"
						viewBox="0 0 24 24"
						stroke-width="2"
						stroke="currentColor"
						fill="none"
						stroke-linecap="round"
						stroke-linejoin="round"
					>
						<path stroke="none" d="M0 0h24v24H0z" fill="none" />
						<path d="M4 5h7" />
						<path d="M7 4c0 4.846 0 7 .5 8" />
						<path
							d="M10 8.5c0 2.286 -2 4.5 -3.5 4.5s-2.5 -1.135 -2.5 -2c0 -2 1 -3 3 -3s5 .57 5 2.857c0 1.524 -.667 2.571 -2 3.143"
						/>
						<path d="M12 20l4 -9l4 9" />
						<path d="M19.1 18h-6.2" />
					</svg>
				</span>
				{#if $page.params.lang != ''}
					<span>
						{jobObj.language} -> {$page.params.lang}
					</span>
				{/if}
			</div>
		{/if}
		Language: {result.language}
	</h2>
	<div class="w-full flex items-center flex-col lg:px-8">
		<video class="max-w-md lg:max-w-2xl" bind:this={video} src={videoSrc} controls />
		<button class="btn btn-info my-2" on:click={save}>Apply changes</button>
		<table class="table w-full mt-3">
			<thead>
				<tr><th>Start</th><th>End</th><th>Subtitle</th></tr>
			</thead>
			<tbody>
				{#each result.segments as seg, i}
					<tr on:click={() => seek(seg.start)}>
						<td
							><input
								class="input"
								type="number"
								bind:value={seg.start}
								min="0"
								step="0.1"
							/></td
						>
						<td
							><input
								class="input"
								type="number"
								bind:value={seg.end}
								min="0"
								step="0.1"
							/></td
						>
						<td><textarea class="textarea w-full" bind:value={seg.text} /></td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
