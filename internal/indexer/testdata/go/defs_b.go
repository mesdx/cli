package sample

import "fmt"

// Config is a duplicate name in a different file for testing.
// This one is a simple type alias.
type Config2 = Config

// FormatString is a helper function.
func FormatString(s string) string {
	return fmt.Sprintf("[%s]", s)
}

// Processor handles data processing.
type Processor struct {
	Name    string
	Workers int
}

// Run starts the processor.
func (p *Processor) Run() error {
	for i := 0; i < p.Workers; i++ {
		fmt.Printf("Worker %d started for %s\n", i, p.Name)
	}
	return nil
}

// Stop halts the processor.
func (p *Processor) Stop() {
	fmt.Println("Stopping", p.Name)
}
