package console

// RunSearchImperative delegates to manager.Search which streams results from sources.
// Kept simple intentionally; can be extended to group/style per-source output.
func (c *ConsoleUI) RunSearchImperative(query string) error { return c.m.Search(query) }
