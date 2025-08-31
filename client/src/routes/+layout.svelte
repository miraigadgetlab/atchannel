<script>
	import "../app.css";
	import TopBar from "$lib/components/TopBar.svelte";
	let { children } = $props();
</script>

<svelte:head>
	<title>channel</title>
	<link rel="icon" type="image/svg+xml" href="/favicon.svg" />
</svelte:head>

<main class="bg-white text-black flex items-center justify-center">
	<div class="h-screen w-300 flex flex-col">
		<TopBar></TopBar>
		{@render children?.()}
	</div>
</main>
