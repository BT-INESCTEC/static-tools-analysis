package tool

var Registry = map[string]Tool{
	"ades":   &Ades{},
	"argus":  &Argus{},
	"zizmor": &Zizmor{},
}
