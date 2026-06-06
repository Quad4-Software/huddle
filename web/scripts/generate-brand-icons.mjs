import { mdiAccountGroup } from '@mdi/js';
import { writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = join(dirname(fileURLToPath(import.meta.url)), '..', 'public');
const path = mdiAccountGroup;

const surface0 = '#0f1117';
const accent = '#7a9e8e';

const landingBox = 48;
const landingIcon = 28;
const landingRadius = 12;
const landingPad = (landingBox - landingIcon) / 2;
const landingScale = landingIcon / 24;

const faviconSvg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${landingBox} ${landingBox}" role="img" aria-label="Huddle">
  <rect width="${landingBox}" height="${landingBox}" rx="${landingRadius}" fill="${surface0}"/>
  <rect width="${landingBox}" height="${landingBox}" rx="${landingRadius}" fill="${accent}" fill-opacity="0.15"/>
  <g transform="translate(${landingPad} ${landingPad}) scale(${landingScale})">
    <path fill="${accent}" d="${path}"/>
  </g>
</svg>
`;

const appSize = 512;
const appRadius = Math.round((landingRadius / landingBox) * appSize);
const appPad = (landingPad / landingBox) * appSize;
const appScale = (landingScale * appSize) / landingBox;

const appIconSvg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${appSize} ${appSize}" role="img" aria-label="Huddle">
  <rect width="${appSize}" height="${appSize}" rx="${appRadius}" fill="${surface0}"/>
  <rect width="${appSize}" height="${appSize}" rx="${appRadius}" fill="${accent}" fill-opacity="0.15"/>
  <g transform="translate(${appPad} ${appPad}) scale(${appScale})">
    <path fill="${accent}" d="${path}"/>
  </g>
</svg>
`;

writeFileSync(join(root, 'favicon.svg'), faviconSvg);
writeFileSync(join(root, 'icon.svg'), appIconSvg);

console.log('Generated favicon.svg and icon.svg from @mdi/js mdiAccountGroup');
