package tools

func Builtins() map[string]Tool {
	m := make(map[string]Tool)
	for _, t := range []Tool{EchoTool(), TimeTool(), HashTool()} {
		m[t.Definition.Name] = t
	}
	return m
}
