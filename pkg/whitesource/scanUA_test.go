package whitesource

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestExecuteUAScan(t *testing.T) {
	t.Parallel()

	t.Run("success - non mta", func(t *testing.T) {
		config := ScanOptions{
			BuildTool:   "maven",
			ProjectName: "test-project",
			ProductName: "test-product",
		}
		utilsMock := NewScanUtilsMock()
		scan := newTestScan(&config)

		err := scan.ExecuteUAScan(&config, utilsMock)
		assert.NoError(t, err)
		assert.Equal(t, "maven", config.BuildTool)
		assert.Contains(t, utilsMock.Calls[1].Params, config.ProductName)
		assert.Contains(t, utilsMock.Calls[1].Params, ".")
	})

	t.Run("success - mta", func(t *testing.T) {
		config := ScanOptions{
			BuildTool:   "mta",
			ProjectName: "test-project",
			ProductName: "test-product",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile("pom.xml", []byte("dummy"))
		utilsMock.AddFile("package.json", []byte(`{"name":"my-module-name"}`))
		scan := newTestScan(&config)

		err := scan.ExecuteUAScan(&config, utilsMock)
		assert.NoError(t, err)
		assert.Equal(t, "mta", config.BuildTool)
		assert.Contains(t, utilsMock.Calls[1].Params, config.ProductName)
		assert.Contains(t, utilsMock.Calls[1].Params, ".")
	})

	t.Run("error - maven", func(t *testing.T) {
		config := ScanOptions{
			AgentDownloadURL: "https://download.ua.org/agent.jar",
			BuildTool:        "mta",
			ProjectName:      "test-project",
			ProductName:      "test-product",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile("pom.xml", []byte("dummy"))
		utilsMock.AddFile("package.json", []byte(`{"name":"my-module-name"}`))
		utilsMock.DownloadError = map[string]error{"https://download.ua.org/agent.jar": fmt.Errorf("failed to download file")}

		scan := newTestScan(&config)
		err := scan.ExecuteUAScan(&config, utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to run scan for maven modules of mta")
	})

	t.Run("error - no pom.xml", func(t *testing.T) {
		config := ScanOptions{
			BuildTool:   "mta",
			ProjectName: "test-project",
			ProductName: "test-product",
		}
		utilsMock := NewScanUtilsMock()
		scan := newTestScan(&config)

		err := scan.ExecuteUAScan(&config, utilsMock)
		assert.EqualError(t, err, "mta project does not contain an aggregator pom.xml in the root - this is mandatory")
	})

	t.Run("error - npm no name", func(t *testing.T) {
		config := ScanOptions{
			BuildTool:   "mta",
			ProjectName: "test-project",
			ProductName: "test-product",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile("pom.xml", []byte("dummy"))
		utilsMock.AddFile("package.json", []byte(`{}`))
		scan := newTestScan(&config)

		err := scan.ExecuteUAScan(&config, utilsMock)
		assert.EqualError(t, err, "failed retrieve project name")
	})

}

func TestExecuteUAScanInPath(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		config := ScanOptions{
			AgentFileName:  "unified-agent.jar",
			ConfigFilePath: "ua.props",
			ProductName:    "test-product",
			ProductVersion: "1",
			ProjectName:    "test-project",
			OrgToken:       "orgTestToken",
			UserToken:      "userTestToken",
			AgentURL:       "https://ws.service.url/agent",
		}
		utilsMock := NewScanUtilsMock()
		scan := newTestScan(&config)

		err := scan.ExecuteUAScanInPath(&config, utilsMock, "")
		assert.NoError(t, err)
		assert.Equal(t, "java", utilsMock.Calls[1].Exec)
		assert.Equal(t, 18, len(utilsMock.Calls[1].Params))
		assert.Contains(t, utilsMock.Calls[1].Params, "-jar")
		assert.Contains(t, utilsMock.Calls[1].Params, "-d")
		assert.Contains(t, utilsMock.Calls[1].Params, ".")
		assert.Contains(t, utilsMock.Calls[1].Params, "-c")
		// name of config file not tested since it is dynamic. This is acceptable here since we test also the size
		assert.Contains(t, utilsMock.Calls[1].Params, "-apiKey")
		assert.Contains(t, utilsMock.Calls[1].Params, config.OrgToken)
		assert.Contains(t, utilsMock.Calls[1].Params, "-userKey")
		assert.Contains(t, utilsMock.Calls[1].Params, config.UserToken)
		assert.Contains(t, utilsMock.Calls[1].Params, "-project")
		assert.Contains(t, utilsMock.Calls[1].Params, config.ProjectName)
		assert.Contains(t, utilsMock.Calls[1].Params, "-product")
		assert.Contains(t, utilsMock.Calls[1].Params, config.ProductName)
		assert.Contains(t, utilsMock.Calls[1].Params, "-productVersion")
		assert.Contains(t, utilsMock.Calls[1].Params, config.ProductVersion)
		assert.Contains(t, utilsMock.Calls[1].Params, "-wss.url")
		assert.Contains(t, utilsMock.Calls[1].Params, config.AgentURL)
	})

	t.Run("success - dedicated path", func(t *testing.T) {
		config := ScanOptions{
			AgentFileName:  "unified-agent.jar",
			ConfigFilePath: "ua.props",
			ProductName:    "test-product",
			ProductVersion: "1",
			ProjectName:    "test-project",
			OrgToken:       "orgTestToken",
			UserToken:      "userTestToken",
			AgentURL:       "https://ws.service.url/agent",
		}
		utilsMock := NewScanUtilsMock()
		scan := newTestScan(&config)

		err := scan.ExecuteUAScanInPath(&config, utilsMock, "./my/test/path")
		assert.NoError(t, err)
		assert.Contains(t, utilsMock.Calls[1].Params, "-d")
		assert.Contains(t, utilsMock.Calls[1].Params, "./my/test/path")
	})

	t.Run("error - download agent", func(t *testing.T) {
		config := ScanOptions{
			AgentDownloadURL: "https://download.ua.org/agent.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.DownloadError = map[string]error{"https://download.ua.org/agent.jar": fmt.Errorf("failed to download file")}
		scan := newTestScan(&config)

		err := scan.ExecuteUAScanInPath(&config, utilsMock, "")
		assert.Contains(t, fmt.Sprint(err), "failed to download unified agent from URL 'https://download.ua.org/agent.jar'")
	})

	t.Run("error - download jre", func(t *testing.T) {
		config := ScanOptions{
			JreDownloadURL: "https://download.jre.org/jvm.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.DownloadError = map[string]error{"https://download.jre.org/jvm.jar": fmt.Errorf("failed to download file")}
		utilsMock.ShouldFailOnCommand = map[string]error{"java": fmt.Errorf("failed to run java")}
		scan := newTestScan(&config)

		err := scan.ExecuteUAScanInPath(&config, utilsMock, "")
		assert.Contains(t, fmt.Sprint(err), "failed to download jre from URL 'https://download.jre.org/jvm.jar'")
	})

	t.Run("error - append scanned projects", func(t *testing.T) {
		config := ScanOptions{}
		utilsMock := NewScanUtilsMock()
		scan := newTestScan(&config)

		err := scan.ExecuteUAScanInPath(&config, utilsMock, "")
		assert.EqualError(t, err, "projectName must not be empty")
	})

	t.Run("error - rewrite config", func(t *testing.T) {
		config := ScanOptions{
			ProjectName: "test-project",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.FileWriteError = fmt.Errorf("failed to write file")
		scan := newTestScan(&config)

		err := scan.ExecuteUAScanInPath(&config, utilsMock, "")
		assert.Contains(t, fmt.Sprint(err), "failed to write file")
	})

	t.Run("error - scan error", func(t *testing.T) {
		config := ScanOptions{
			ProjectName: "test-project",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.ShouldFailOnCommand = map[string]error{
			"java": fmt.Errorf("failed to run java"),
		}
		scan := newTestScan(&config)

		err := scan.ExecuteUAScanInPath(&config, utilsMock, "")
		assert.Contains(t, fmt.Sprint(err), "failed to execute WhiteSource scan with exit code")
	})
}

func TestEvaluateExitCode(t *testing.T) {
	tt := []struct {
		exitCode int
		expected log.ErrorCategory
	}{
		{exitCode: 255, expected: log.ErrorUndefined},
		{exitCode: 254, expected: log.ErrorCompliance},
		{exitCode: 253, expected: log.ErrorUndefined},
		{exitCode: 252, expected: log.ErrorInfrastructure},
		{exitCode: 251, expected: log.ErrorService},
		{exitCode: 250, expected: log.ErrorCustom},
		{exitCode: 200, expected: log.ErrorUndefined},
	}

	for _, test := range tt {
		evaluateExitCode(test.exitCode)
		assert.Equal(t, test.expected, log.GetErrorCategory(), fmt.Sprintf("test for exit code %v failed", test.exitCode))
	}
}

func TestDownloadAgent(t *testing.T) {
	t.Parallel()

	t.Run("success - download", func(t *testing.T) {
		config := ScanOptions{
			AgentDownloadURL: "https://download.ua.org/agent.jar",
			AgentFileName:    "unified-agent.jar",
		}
		utilsMock := NewScanUtilsMock()

		err := downloadAgent(&config, utilsMock)
		assert.NoError(t, err, "error occured although none expected")
		assert.Len(t, utilsMock.DownloadedFiles, 1)
		assert.Equal(t, "https://download.ua.org/agent.jar", utilsMock.DownloadedFiles[0].SourceURL)
		assert.Equal(t, "unified-agent.jar", utilsMock.DownloadedFiles[0].FilePath)
	})

	t.Run("success - no download", func(t *testing.T) {
		config := ScanOptions{
			AgentDownloadURL: "https://download.ua.org/agent.jar",
			AgentFileName:    "unified-agent.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile("unified-agent.jar", []byte("dummy"))

		err := downloadAgent(&config, utilsMock)
		assert.NoError(t, err, "error occured although none expected")
		assert.Len(t, utilsMock.DownloadedFiles, 0)
	})

	t.Run("error - file existence", func(t *testing.T) {
		config := ScanOptions{
			AgentDownloadURL: "https://download.ua.org/agent.jar",
			AgentFileName:    "unified-agent.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.FileExistsErrors = map[string]error{"unified-agent.jar": fmt.Errorf("failed to check existence")}

		err := downloadAgent(&config, utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to check if file 'unified-agent.jar' exists")
	})

	t.Run("error - download", func(t *testing.T) {
		config := ScanOptions{
			AgentDownloadURL: "https://download.ua.org/agent.jar",
			AgentFileName:    "unified-agent.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.DownloadError = map[string]error{"https://download.ua.org/agent.jar": fmt.Errorf("failed to download file")}

		err := downloadAgent(&config, utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to download unified agent from URL")
	})
}

func TestDownloadJre(t *testing.T) {
	t.Parallel()

	t.Run("success - no download required", func(t *testing.T) {
		config := ScanOptions{
			JreDownloadURL: "https://download.jre.org/jvm.jar",
		}
		utilsMock := NewScanUtilsMock()

		jre, err := downloadJre(&config, utilsMock)
		assert.NoError(t, err)
		assert.Equal(t, "java", jre)
		assert.Equal(t, "java", utilsMock.Calls[0].Exec)
		assert.Equal(t, []string{"-version"}, utilsMock.Calls[0].Params)
	})

	t.Run("success - previously downloaded", func(t *testing.T) {
		config := ScanOptions{
			JreDownloadURL: "https://download.jre.org/jvm.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile(filepath.Join(jvmDir, "bin", "java"), []byte("dummy"))

		jre, err := downloadJre(&config, utilsMock)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(jvmDir, "bin", "java"), jre)
	})

	t.Run("success - jre downloaded", func(t *testing.T) {
		config := ScanOptions{
			JreDownloadURL: "https://download.jre.org/jvm.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.ShouldFailOnCommand = map[string]error{"java": fmt.Errorf("failed to run java")}

		jre, err := downloadJre(&config, utilsMock)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(jvmDir, "bin", "java"), jre)
		assert.Equal(t, "https://download.jre.org/jvm.jar", utilsMock.DownloadedFiles[0].SourceURL)
		exists, _ := utilsMock.DirExists(jvmDir)
		assert.True(t, exists)
		assert.Equal(t, "tar", utilsMock.Calls[1].Exec)
		assert.Equal(t, fmt.Sprintf("--directory=%v", jvmDir), utilsMock.Calls[1].Params[0])
	})

	t.Run("error - download", func(t *testing.T) {
		config := ScanOptions{
			JreDownloadURL: "https://download.jre.org/jvm.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.ShouldFailOnCommand = map[string]error{"java": fmt.Errorf("failed to run java")}
		utilsMock.DownloadError = map[string]error{"https://download.jre.org/jvm.jar": fmt.Errorf("failed to download file")}

		_, err := downloadJre(&config, utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to download jre from URL")
	})

	t.Run("error - tar execution", func(t *testing.T) {
		config := ScanOptions{
			JreDownloadURL: "https://download.jre.org/jvm.jar",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.ShouldFailOnCommand = map[string]error{
			"java": fmt.Errorf("failed to run java"),
			"tar":  fmt.Errorf("failed to run tar"),
		}
		_, err := downloadJre(&config, utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to extract")
	})
}

func TestRemoveJre(t *testing.T) {
	t.Parallel()

	t.Run("success - no removal required", func(t *testing.T) {
		utilsMock := NewScanUtilsMock()
		err := removeJre("java", utilsMock)
		assert.NoError(t, err, "error occured although none expected")
	})

	t.Run("success - with removal", func(t *testing.T) {
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile(jvmTarGz, []byte("dummy"))
		err := removeJre("./jvm/bin/java", utilsMock)
		assert.NoError(t, err, "error occured although none expected")
		assert.Contains(t, utilsMock.RemoveAllDirs, jvmDir)
		assert.True(t, utilsMock.HasRemovedFile(jvmTarGz))
	})

	t.Run("error - remove jvm directory", func(t *testing.T) {
		utilsMock := NewScanUtilsMock()
		utilsMock.RemoveAllError = map[string]error{jvmDir: fmt.Errorf("failed to remove directory")}

		err := removeJre("./jvm/bin/java", utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to remove downloaded and extracted jvm")
	})

	t.Run("error - remove jvm tar.gz", func(t *testing.T) {
		utilsMock := NewScanUtilsMock()

		err := removeJre("./jvm/bin/java", utilsMock)
		assert.Contains(t, fmt.Sprint(err), "failed to remove downloaded")
	})
}
