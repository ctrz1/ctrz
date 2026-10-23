package fs

type Manager struct {
	Utils map[string]bool
}

func New() Manager {
	return Manager{
		Utils: getUtils(),
	}
}
