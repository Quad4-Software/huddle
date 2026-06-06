import { render } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import { mdiMicrophone } from '../icons';
import Icon from './Icon.svelte';

describe('Icon', () => {
  it('renders the provided mdi path', () => {
    const { container } = render(Icon, { path: mdiMicrophone, size: 24 });
    const path = container.querySelector('path');
    expect(path?.getAttribute('d')).toBe(mdiMicrophone);
    expect(container.querySelector('svg')?.getAttribute('width')).toBe('24');
  });
});
