package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Metadata struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
}

type Item struct {
	Name     string   `json:"name"`
	Metadata Metadata `json:"metadata"`
}

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) SayHello() {
	runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Title:   "Hello",
		Message: "Hello from the menu!",
	})
}

func (a *App) GetItem() Item {
	return Item{
		Name: "Table",
		Metadata: Metadata{
			ID:    '1',
			Label: "Bed-Table",
		},
	}
}

func (a *App) GetItems() []Item {
	return []Item{
		{
			Name: "Table",
			Metadata: Metadata{
				ID:    1,
				Label: "Study Table",
			},
		},
		{
			Name: "HeadPhones",
			Metadata: Metadata{
				ID:    2,
				Label: "HeadPhones with Mic",
			},
		},
	}
}
