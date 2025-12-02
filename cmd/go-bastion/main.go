package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const templateRepoURL = "https://github.com/AlejandroMBJS/goBastion.git"
const originalModuleName = "go-native-fastapi"
const modulePrefix = "github.com/AlejandroMBJS/"

func main() {
	var projectName string
	if len(os.Args) > 1 {
		projectName = os.Args[1]
	} else {
		fmt.Print("¿Cómo quieres llamar a tu nuevo proyecto? > ")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("❌ Error leyendo el nombre del proyecto: %v", err)
		}
		projectName = strings.TrimSpace(line)
	}

	if projectName == "" {
		log.Fatal("❌ El nombre del proyecto no puede estar vacío.")
	}

	runGenerator(projectName)
}

func runGenerator(projectName string) {
	targetDir := "./" + projectName
	fmt.Printf("⚙ Creando nuevo proyecto '%s' en %s\n", projectName, targetDir)

	cloneTemplateRepo(targetDir)
	removeGoBastionDir(targetDir)
	removeGitDir(targetDir)
	newModuleName := buildNewModuleName(projectName)
	replaceModuleNameInFiles(targetDir, originalModuleName, newModuleName)
	addReplaceDirective(targetDir, newModuleName)
	runGoModTidy(targetDir)

	fmt.Printf("✅ Proyecto creado exitosamente.\n")
	fmt.Printf("➡ Ahora puedes entrar a la carpeta: cd %s\n", projectName)
}

func cloneTemplateRepo(targetDir string) {
	fmt.Printf("Clonando el repositorio template en %s...\n", targetDir)
	cmd := exec.Command("git", "clone", templateRepoURL, targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("❌ Error clonando el repositorio: %v", err)
	}
}

func removeGitDir(targetDir string) {
	fmt.Println("Eliminando el directorio .git...")
	if err := os.RemoveAll(filepath.Join(targetDir, ".git")); err != nil {
		log.Fatalf("❌ Error eliminando el directorio .git: %v", err)
	}
}

func removeGoBastionDir(targetDir string) {
	fmt.Println("Eliminando el directorio cmd/go-bastion...")
	if err := os.RemoveAll(filepath.Join(targetDir, "cmd/go-bastion")); err != nil {
		log.Fatalf("❌ Error eliminando el directorio cmd/go-bastion: %v", err)
	}
}

func buildNewModuleName(projectName string) string {
	if modulePrefix == "" {
		return projectName
	}
	return modulePrefix + projectName
}

func replaceModuleNameInFiles(targetDir, oldModule, newModule string) {
	fmt.Printf("Reemplazando el nombre del módulo '%s' por '%s'...\n", oldModule, newModule)
	err := filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && shouldProcessFile(path) {
			if err := replaceInFile(path, oldModule, newModule); err != nil {
				log.Printf("⚠️  No se pudo reemplazar en el archivo %s: %v", path, err)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("❌ Error recorriendo los archivos: %v", err)
	}
}

func shouldProcessFile(path string) bool {
	return strings.HasSuffix(path, ".go") || filepath.Base(path) == "go.mod"
}

func replaceInFile(path, old, new string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	newContent := strings.ReplaceAll(string(content), old, new)

	if string(content) != newContent {
		return os.WriteFile(path, []byte(newContent), 0644)
	}
	return nil
}

func runGoModTidy(targetDir string) {
	fmt.Println("Ejecutando 'go mod tidy'...")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = targetDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("❌ Error ejecutando 'go mod tidy': %v", err)
	}
}

func addReplaceDirective(targetDir, newModule string) {
	fmt.Println("Añadiendo directiva 'replace' al go.mod...")
	goModPath := filepath.Join(targetDir, "go.mod")
	f, err := os.OpenFile(goModPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("❌ Error abriendo go.mod: %v", err)
	}
	defer f.Close()

	replaceDirective := fmt.Sprintf("\nreplace %s => .\n", newModule)
	if _, err := f.WriteString(replaceDirective); err != nil {
		log.Fatalf("❌ Error escribiendo en go.mod: %v", err)
	}
}