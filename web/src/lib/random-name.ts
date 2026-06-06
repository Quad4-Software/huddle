const adjectives = [
  'Swift',
  'Calm',
  'Bold',
  'Quiet',
  'Bright',
  'Clever',
  'Lucky',
  'Merry',
  'Noble',
  'Quick',
  'Sunny',
  'Witty',
  'Brave',
  'Cosmic',
  'Dapper',
  'Eager',
  'Gentle',
  'Happy',
  'Jolly',
  'Keen',
];

const nouns = [
  'Fox',
  'River',
  'Comet',
  'Falcon',
  'Harbor',
  'Maple',
  'Panda',
  'Quartz',
  'Robin',
  'Spruce',
  'Tiger',
  'Willow',
  'Badger',
  'Cedar',
  'Drift',
  'Ember',
  'Finch',
  'Grove',
  'Heron',
  'Lynx',
];

export function randomDisplayName(): string {
  const adj = adjectives[Math.floor(Math.random() * adjectives.length)];
  const noun = nouns[Math.floor(Math.random() * nouns.length)];
  return `${adj} ${noun}`;
}

export function randomRoomName(): string {
  return randomDisplayName();
}
