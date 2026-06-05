package player

type Player struct {
	progressMs int
	playing    bool
}

func New() *Player {
	return &Player{}
}

func (p *Player) ProgressMs() int {
	return p.progressMs
}

func (p *Player) SetNowPlaying(np NowPlayingEntry) {
	p.progressMs = np.ProgressMs
	p.playing = np.Playing
}

func (p *Player) Tick() {
	if p.playing {
		p.progressMs += 1000
	}
}
