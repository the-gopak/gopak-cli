package console

import (
	"fmt"

	survey "github.com/AlecAivazis/survey/v2"
)

func (c *ConsoleUI) RunRemoveImperative(name string) error {
	ok := false
	if err := survey.AskOne(&survey.Confirm{Message: messageRemoveConfirm(name), Default: true}, &ok); err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if err := c.m.Remove(name); err != nil {
		return err
	}
	fmt.Println("removed:", name)
	return nil
}

func messageRemoveConfirm(name string) string { return fmt.Sprintf("Remove %s?", name) }
