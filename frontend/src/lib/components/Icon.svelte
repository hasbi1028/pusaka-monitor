<script>
  import { paths } from './iconPaths.js';

  let { name, class: cls = '', spin = false } = $props();

  let entry = $derived(paths[name]);
</script>

{#if entry}
  <svg
    class="fa-icon {cls}"
    class:fa-spin={spin}
    viewBox={entry[0]}
    fill="currentColor"
    aria-hidden="true"
    focusable="false"
  ><path d={entry[1]} /></svg>
{/if}

<style>
  .fa-icon {
    display: inline-block;
    width: 1em;
    height: 1em;
    flex-shrink: 0;
    vertical-align: -0.125em;
  }
  .fa-spin {
    animation: fa-spin 1s linear infinite;
  }
  @keyframes fa-spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }
</style>
