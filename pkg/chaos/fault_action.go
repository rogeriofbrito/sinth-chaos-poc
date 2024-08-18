package chaos

type IFaultAction interface {
	Inject(pid int, params interface{}) error
	Remove(pid int, params interface{}) error
}
