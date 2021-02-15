package helpers

type STATE struct {
	Current string
	Time    int
}

func (s STATE) Set(newState string, time int) {
	s.Current = newState
	s.Time = time
}
