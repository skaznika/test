<script>
  import "../app.postcss";
  import { SvelteToast } from '@zerodevx/svelte-toast'
</script>

<SvelteToast />

<div class="navbar bg-transparent px-8 mb-4">
    <div class="flex-1">
      <a href="/" class="btn btn-ghost normal-case text-xl">WebWhisper+</a>
    </div>
    <div class="flex-none hidden">
      <ul class="menu menu-horizontal px-1">
        <li><a href="/" class="hidden">Item 1</a></li>
      </ul>
    </div>
  </div>

<slot></slot>