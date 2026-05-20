package docker

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"hacklab/internal/lab"
)

// Manager handles Docker operations for labs via the docker CLI
type Manager struct {
	labDir  string
	labName string
}

// NewManager creates a new Docker manager
func NewManager(labDir, labName string) (*Manager, error) {
	// Check if docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		return nil, fmt.Errorf("docker not found in PATH — install docker first")
	}

	// Check if docker daemon is running
	cmd := exec.Command("docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("docker daemon not running — start it first")
	}

	return &Manager{
		labDir:  labDir,
		labName: labName,
	}, nil
}

// StartSingle starts a single container lab
func (m *Manager) StartSingle(mf *lab.Manifest) (string, int, error) {
	containerName := fmt.Sprintf("hacklab-%s", m.labName)

	// Check if already running
	if m.isRunning(containerName) {
		return "", 0, fmt.Errorf("lab '%s' is already running (stop it first with 'hacklab stop %s')", m.labName, m.labName)
	}

	// Remove stale container
	m.removeContainer(containerName)

	// Pull image
	fmt.Printf("  📦 Pulling image %s...\n", mf.Image)
	cmd := exec.Command("docker", "pull", mf.Image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", 0, fmt.Errorf("pulling image: %w", err)
	}

	// Build port bindings
	hostPort := mf.Port
	portSpec := fmt.Sprintf("127.0.0.1:%d:%d/tcp", hostPort, hostPort)

	// Create and run container
	fmt.Printf("  🚀 Starting container...\n")
	cmd = exec.Command("docker", "run", "-d",
		"--name", containerName,
		"-p", portSpec,
		"--memory", "512m",
		"--cpus", "1",
		"--label", fmt.Sprintf("hacklab.lab=%s", m.labName),
		"-e", fmt.Sprintf("HACKLAB_LAB=%s", m.labName),
		mf.Image,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", 0, fmt.Errorf("starting container: %w", err)
	}

	return containerName, hostPort, nil
}

// StartCompose starts a docker-compose lab
func (m *Manager) StartCompose(mf *lab.Manifest) error {
	projectName := fmt.Sprintf("hacklab-%s", m.labName)
	composePath := filepath.Join(m.labDir, mf.ComposeFile)

	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return fmt.Errorf("compose file not found: %s", composePath)
	}

	// Check if already running
	cmd := exec.Command("docker", "compose", "-p", projectName, "ps", "--services", "--filter", "status=running")
	cmd.Dir = m.labDir
	output, err := cmd.Output()
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {
		return fmt.Errorf("lab '%s' is already running (stop it first with 'hacklab stop %s')", m.labName, m.labName)
	}

	// Start with docker compose
	fmt.Printf("  📦 Starting docker-compose lab...\n")
	cmd = exec.Command("docker", "compose", "-p", projectName, "up", "-d")
	cmd.Dir = m.labDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose up: %w", err)
	}

	return nil
}

// WaitForReady polls a URL until it responds or times out
func (m *Manager) WaitForReady(url string, timeoutSecs int) error {
	if url == "" {
		time.Sleep(2 * time.Second)
		return nil
	}

	fmt.Printf("  ⏳ Waiting for lab to be ready (max %ds)...\n", timeoutSecs)

	httpClient := &http.Client{Timeout: 3 * time.Second}
	deadline := time.Now().Add(time.Duration(timeoutSecs) * time.Second)

	for time.Now().Before(deadline) {
		resp, err := httpClient.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				return nil
			}
		}
		time.Sleep(1 * time.Second)
		fmt.Print(".")
	}

	return fmt.Errorf("timed out waiting for %s", url)
}

// Stop stops a lab by name
func Stop(labName string) error {
	containerName := fmt.Sprintf("hacklab-%s", labName)

	// Try docker compose first
	cmd := exec.Command("docker", "compose", "-p", containerName, "down", "-v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err == nil {
		fmt.Printf("  ✅ Lab '%s' stopped\n", labName)
		return nil
	}

	// Fallback: stop single container
	stopCmd := exec.Command("docker", "stop", containerName)
	stopCmd.Stdout = nil
	stopCmd.Stderr = nil
	if err := stopCmd.Run(); err != nil {
		fmt.Printf("  ℹ️  No running lab '%s' found\n", labName)
		return nil
	}

	rmCmd := exec.Command("docker", "rm", containerName)
	rmCmd.Stdout = nil
	rmCmd.Stderr = nil
	_ = rmCmd.Run()

	fmt.Printf("  ✅ Lab '%s' stopped\n", labName)
	return nil
}

// ListRunning shows all hacklab containers
func ListRunning() ([]string, error) {
	cmd := exec.Command("docker", "ps",
		"--filter", "label=hacklab.lab",
		"--format", "{{.Label \"hacklab.lab\"}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var labs []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			labs = append(labs, line)
		}
	}
	return labs, nil
}

// isRunning checks if a container is running
func (m *Manager) isRunning(name string) bool {
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", name)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

// removeContainer removes a container if it exists
func (m *Manager) removeContainer(name string) {
	cmd := exec.Command("docker", "rm", "-f", name)
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Run()
}
