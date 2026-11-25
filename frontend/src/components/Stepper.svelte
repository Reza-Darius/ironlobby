<script lang="ts">
  import Icon from './Icon.svelte'

  let {
    value,
    min = 1,
    max = 32,
    disabled = false,
    label = 'slots',
    onchange,
  }: {
    value: number
    min?: number
    max?: number
    disabled?: boolean
    label?: string
    onchange: (v: number) => void
  } = $props()
</script>

<div class="stepper" aria-label={label}>
  <button
    type="button"
    class="btn icon sm"
    aria-label="Fewer {label}"
    disabled={disabled || value <= min}
    onclick={() => onchange(value - 1)}
  >
    <Icon name="minus" size={14} />
  </button>
  <span class="value" aria-live="polite">{value}</span>
  <button
    type="button"
    class="btn icon sm"
    aria-label="More {label}"
    disabled={disabled || value >= max}
    onclick={() => onchange(value + 1)}
  >
    <Icon name="plus" size={14} />
  </button>
</div>

<style>
  .stepper {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    background: var(--bg-2);
    border: 1px solid var(--line);
    border-radius: var(--radius);
    padding: 2px;
  }
  .value {
    min-width: 26px;
    text-align: center;
    font-family: var(--mono);
    font-weight: 600;
    font-size: 14px;
  }
  .stepper :global(.btn) {
    border-color: transparent;
    background: transparent;
  }
</style>
