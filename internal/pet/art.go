package pet

var frames = map[Mood]string{
	MoodHappy: `
  (◕‿◕)
  /|  |\
   |  |
`,
	MoodNeutral: `
  (•_•)
  /|  |\
   |  |
`,
	MoodSad: `
  (╥_╥)
  /|  |\
   |  |
`,
}

func Art(m Mood) string {
	return frames[m]
}
