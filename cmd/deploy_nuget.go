package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
  "runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	dryRun    bool
	searchDir string
)

type DeploymentProgress struct {
	Bar       *progressbar.ProgressBar
	ProjectMu sync.Mutex
}

type Project struct {
	Name string
	Path string
	Dir  string
}

type model struct {
	projects []Project
	cursor   int
	selected map[int]bool
	done     bool
}

var (
	lavender = lipgloss.NewStyle().Foreground(lipgloss.Color("#b7bdf8"))
	yellow   = lipgloss.NewStyle().Foreground(lipgloss.Color("#eed49f"))
	mauve    = lipgloss.NewStyle().Foreground(lipgloss.Color("#c6a0f6"))
)

var(
  listTitle = "\nselect nuget packages to deploy (space to select, enter to confirm):\n\n" 
  quitMsg = "\nPress q to quit.\n"
)

func (m model) Init() tea.Cmd { return nil }

func (m model) View() string {
	s := listTitle
	for i, project := range m.projects {
		cursor := " "
		if m.cursor == i {
			cursor = mauve.Render(">")
		}
		checked := " "
		if m.selected[i] {
			checked = yellow.Render("x")
		}
		projectInfo := fmt.Sprintf("%s", project.Name)
		if m.cursor == i {
			projectInfo = mauve.Render(projectInfo)
		}

		s += fmt.Sprintf("%s [%s] %s \n", cursor, checked, projectInfo)
	}
	s += quitMsg
	return s
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.projects)-1 {
				m.cursor++
			}
		case " ":
			m.selected[m.cursor] = !m.selected[m.cursor]
		case "enter":
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func newProgressBar(total int, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions(total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(30),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: "-",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetPredictTime(true),
	)
}

func findSvPackagesDir() (string, error) {
	// Convert to absolute path to handle relative paths properly
	absPath, err := filepath.Abs(".")
	if err != nil {
		return "", fmt.Errorf("error getting absolute path: %v", err)
	}

	// First check if current directory is Sv.Packages
	currentInfo, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("error accessing current directory: %v", err)
	}

	if currentInfo.IsDir() && strings.EqualFold(currentInfo.Name(), "Sv.Packages") {
		return absPath, nil
	}

	// Walk up through parent directories
	for dir := absPath; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
		svPackagesPath := filepath.Join(dir, "Sv.Packages")
		info, err := os.Stat(svPackagesPath)
		if err == nil && info.IsDir() {
			return svPackagesPath, nil
		}
	}

	return "", fmt.Errorf("Sv.Packages directory not found in current directory or any parent directory")
}

func findProjects() ([]Project, error) {
	// Find Sv.Packages directory
	svPackagesPath, err := findSvPackagesDir()
	if err != nil {
		return nil, err
	}

	var projects []Project
	err = filepath.Walk(svPackagesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".csproj") {
			dir := filepath.Dir(path)
			// Skip the Sv.Packages.csproj itself
			if strings.EqualFold(filepath.Base(dir), "Sv.Packages") {
				return nil
			}
			projects = append(projects, Project{
				Name: filepath.Base(dir),
				Path: path,
				Dir:  dir,
			})
		}
		return nil
	})

	if len(projects) == 0 {
		return nil, fmt.Errorf("no projects found in Sv.Packages directory")
	}

	return projects, nil
}

func cleanPackages(projectDir string) error {
    nupkgPath := filepath.Join(projectDir, "bin", "Release", "*.nupkg")
    if runtime.GOOS == "windows" {
        removeCmd := exec.Command("cmd", "/c", "del", "/f", "/q", nupkgPath)
        return removeCmd.Run()
    }
    // Unix-like systems
    removeCmd := exec.Command("rm", "-f", nupkgPath)
    return removeCmd.Run()
}

func deployProject(project Project, dryRun bool) error {
	// Create a step-by-step progress bar for this project
	steps := []string{
		"Cleaning packages",
		"Building project",
		"Publishing NuGet package",
	}
	bar := newProgressBar(len(steps), fmt.Sprintf("Deploying %s", project.Name))

	log.Printf("\nProcessing: %s\n", yellow.Render(project.Name))
	log.Printf("Directory: %s\n", project.Dir)

	if dryRun {
		log.Printf("[DRY RUN] Would deploy: %s\n", project.Path)
		bar.Add(len(steps)) // Complete the bar for dry run
		return nil
	}

	if err := os.Chdir(project.Dir); err != nil {
		return fmt.Errorf("failed to change directory: %v", err)
	}

	// Clean packages
	bar.Describe("Cleaning packages")
	nupkgPath := filepath.Join(project.Dir, "bin", "Release", "*.nupkg")
	removeCmd := exec.Command("del", "/f", "/q", nupkgPath)
	removeCmd.Run()
	bar.Add(1)

	// Build
	bar.Describe("Building project")
	msbuildPath := `C:\Program Files\Microsoft Visual Studio\2022\Professional\Common7\Tools\VsDevCmd.bat`
	buildCmd := exec.Command("cmd", "/c",
		fmt.Sprintf(`"%s" && msbuild.exe "%s.csproj" /property:Configuration=Release /t:Rebuild`,
			msbuildPath, project.Name))
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("build failed: %v", err)
	}
	bar.Add(1)

	var response string
	fmt.Print("\nDid the build succeed? (y/n): ")
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" {
		return fmt.Errorf("build reported as failed by user")
	}

	// Push package
	bar.Describe("Publishing NuGet package")
	pushCmd := exec.Command("dotnet", "nuget", "push",
		"--source", "Silvervine",
		"--api-key", "az",
		"--skip-duplicate",
		filepath.Join("bin", "Release", fmt.Sprintf("%s*.nupkg", project.Name)))
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr

	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("push failed: %v", err)
	}
	bar.Add(1)

	fmt.Print("\nDid the package deploy successfully? (y/n): ")
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" {
		return fmt.Errorf("deployment reported as failed by user")
	}

	return nil
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy NuGet packages",
	Run:   deployCommand,
}

func init() {
	deployCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print commands without executing them")
	RootCmd.AddCommand(deployCmd)
}

func deployCommand(_ *cobra.Command, _ []string) {

	projects, err := findProjects()

	if err != nil {
		log.Fatalf("Failed to find Nuget projects: %v\n", err)
	}

	m := model{
		projects: projects,
		selected: make(map[int]bool),
	}

	program := tea.NewProgram(m)
	finalModel, err := program.Run()

	if err != nil {
		log.Fatalf("Error running program: %v\n", err)
	}

	finalState := finalModel.(model)

	hasSelected := false
	for _, selected := range finalState.selected {
		if selected {
			hasSelected = true
			break
		}
	}

	if !hasSelected {
		log.Printf("No nuget packages selected")
		return
	}

	// Count selected projects
	selectedCount := 0
	for i := range finalState.projects {
		if finalState.selected[i] {
			selectedCount++
		}
	}

	// Create overall progress bar
	overallBar := newProgressBar(selectedCount, "Overall Progress")

	for i, project := range finalState.projects {
		if !finalState.selected[i] {
			continue
		}

		err := deployProject(project, dryRun)
		if err != nil {
			log.Printf("Error deploying %s: %v\n", project.Name, err)
			fmt.Print("Continue with remaining projects? (y/n): ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				break
			}
		} else {
			log.Printf("Successfully deployed %s!\n", project.Name)
		}
		overallBar.Add(1)
	}
}
