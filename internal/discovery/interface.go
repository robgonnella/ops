package discovery

type Service interface {
	SetTargets(targets []string)
	MonitorNetwork()
	Stop()
}
