package steps

import (
	"os"
	"testing"

	"gitlab.cee.redhat.com/clobrano/ccoctl-sso/pkg/config"
	"gitlab.cee.redhat.com/clobrano/ccoctl-sso/pkg/logger"
	"gitlab.cee.redhat.com/clobrano/ccoctl-sso/pkg/util"
)

func TestStep6CreateAWSResources(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	cfg := &config.Config{
		ReleaseImage: "quay.io/test:4.12.0-x86_64",
		ClusterName:  "test-cluster",
		AwsRegion:    "us-east-2",
		OutputDir:    "_output",
	}
	log := logger.New(logger.LevelQuiet, nil)
	executor := util.NewMockExecutor()

	// Create required directories
	os.MkdirAll("artifacts/4.12.0-x86_64/bin", 0755)
	os.MkdirAll("artifacts/4.12.0-x86_64/credreqs", 0755)

	step, err := NewStep6(cfg, log, executor)
	if err != nil {
		t.Fatalf("Failed to create step: %v", err)
	}

	err = step.Execute()
	if err != nil {
		t.Fatalf("Step execution failed: %v", err)
	}

	if !executor.WasExecutedContaining("ccoctl") {
		t.Error("Expected ccoctl command to be executed")
	}
	if !executor.WasExecutedContaining("aws create-all") {
		t.Error("Expected 'aws create-all' in command")
	}
	if !executor.WasExecutedContaining("--name test-cluster") {
		t.Error("Expected cluster name in command")
	}
}

func TestStep6WithPrivateBucket(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	cfg := &config.Config{
		ReleaseImage:  "quay.io/test:4.12.0-x86_64",
		ClusterName:   "test-cluster",
		AwsRegion:     "us-east-2",
		OutputDir:     "_output",
		PrivateBucket: true,
	}
	log := logger.New(logger.LevelQuiet, nil)
	executor := util.NewMockExecutor()

	os.MkdirAll("artifacts/4.12.0-x86_64/bin", 0755)
	os.MkdirAll("artifacts/4.12.0-x86_64/credreqs", 0755)

	step, err := NewStep6(cfg, log, executor)
	if err != nil {
		t.Fatalf("Failed to create step: %v", err)
	}

	err = step.Execute()
	if err != nil {
		t.Fatalf("Step execution failed: %v", err)
	}

	if !executor.WasExecutedContaining("--create-private-s3-bucket") {
		t.Error("Expected '--create-private-s3-bucket' flag when PrivateBucket is true")
	}
}

func TestStep7CopyManifests(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	cfg := &config.Config{
		ReleaseImage: "quay.io/test:4.12.0-x86_64",
		OutputDir:    "_output",
	}
	log := logger.New(logger.LevelQuiet, nil)
	executor := util.NewMockExecutor()

	// Create source directory with files
	os.MkdirAll("_output/manifests", 0755)
	os.WriteFile("_output/manifests/test.yaml", []byte("test content"), 0644)

	step, err := NewStep7(cfg, log, executor)
	if err != nil {
		t.Fatalf("Failed to create step: %v", err)
	}

	err = step.Execute()
	if err != nil {
		t.Fatalf("Step execution failed: %v", err)
	}

	// Verify files were copied
	if !util.FileExists("manifests/test.yaml") {
		t.Error("Manifest file was not copied")
	}
}

func TestStep8CopyTLS(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	cfg := &config.Config{
		ReleaseImage: "quay.io/test:4.12.0-x86_64",
		OutputDir:    "_output",
	}
	log := logger.New(logger.LevelQuiet, nil)
	executor := util.NewMockExecutor()

	// Create source directory with files
	os.MkdirAll("_output/tls", 0755)
	os.WriteFile("_output/tls/ca.pem", []byte("cert content"), 0644)

	step, err := NewStep8(cfg, log, executor)
	if err != nil {
		t.Fatalf("Failed to create step: %v", err)
	}

	err = step.Execute()
	if err != nil {
		t.Fatalf("Step execution failed: %v", err)
	}

	// Verify files were copied
	if !util.FileExists("tls/ca.pem") {
		t.Error("TLS file was not copied")
	}
}

func TestStep9DeployCluster(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	cfg := &config.Config{
		ReleaseImage: "quay.io/test:4.12.0-x86_64",
	}
	log := logger.New(logger.LevelQuiet, nil)
	executor := util.NewMockExecutor()

	os.MkdirAll("artifacts/4.12.0-x86_64/bin", 0755)

	step, err := NewStep9(cfg, log, executor)
	if err != nil {
		t.Fatalf("Failed to create step: %v", err)
	}

	err = step.Execute()
	if err != nil {
		t.Fatalf("Step execution failed: %v", err)
	}

	if !executor.WasExecutedContaining("openshift-install") {
		t.Error("Expected openshift-install command")
	}
	if !executor.WasExecutedContaining("create cluster") {
		t.Error("Expected 'create cluster' in command")
	}
}

func TestStep10Verify(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	cfg := &config.Config{
		ReleaseImage: "quay.io/test:4.12.0-x86_64",
	}
	log := logger.New(logger.LevelQuiet, nil)
	executor := util.NewMockExecutor()

	// Mock verification commands
	executor.SetError("oc get secrets -n kube-system aws-creds",
		os.ErrNotExist) // Root creds should not exist
	executor.SetOutput("oc get secrets -n openshift-image-registry installer-cloud-credentials -o json",
		`{"data":{"credentials":"role_arn = arn:aws:iam::123456789:role/test\nweb_identity_token_file = /var/run/secrets/token"}}`)

	step, err := NewStep10(cfg, log, executor)
	if err != nil {
		t.Fatalf("Failed to create step: %v", err)
	}

	err = step.Execute()
	if err != nil {
		t.Fatalf("Step execution failed: %v", err)
	}

	// Verify both checks were executed
	if !executor.WasExecutedContaining("kube-system") {
		t.Error("Expected root credentials check")
	}
	if !executor.WasExecutedContaining("openshift-image-registry") {
		t.Error("Expected IAM role check")
	}
}
