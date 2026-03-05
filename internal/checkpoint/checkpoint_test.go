package checkpoint

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// createTestCheckpoints is a helper that creates checkpoint files in a temp dir
// and returns the base dir and profile dir.
func createTestCheckpoints(t *testing.T, profile string, filenames []string) string {
	t.Helper()
	baseDir := t.TempDir()
	profileDir := filepath.Join(baseDir, profile)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		t.Fatalf("failed to create profile dir: %v", err)
	}
	for _, f := range filenames {
		if err := os.WriteFile(filepath.Join(profileDir, f), []byte("-- sql"), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", f, err)
		}
	}
	return baseDir
}

func TestListCheckpointFilenames(t *testing.T) {
	t.Run("lists existing checkpoints", func(t *testing.T) {
		files := []string{"checkpoint_1.sql", "checkpoint_2.sql", "checkpoint_3.sql"}
		baseDir := createTestCheckpoints(t, "default", files)

		got, err := ListCheckpointFilenames(baseDir, "default")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sort.Strings(got)
		sort.Strings(files)
		if len(got) != len(files) {
			t.Fatalf("got %d files, want %d", len(got), len(files))
		}
		for i := range got {
			if got[i] != files[i] {
				t.Errorf("got %s, want %s", got[i], files[i])
			}
		}
	})

	t.Run("returns empty for no checkpoints", func(t *testing.T) {
		baseDir := createTestCheckpoints(t, "default", nil)

		got, err := ListCheckpointFilenames(baseDir, "default")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("got %d files, want 0", len(got))
		}
	})

	t.Run("ignores non-checkpoint files", func(t *testing.T) {
		baseDir := createTestCheckpoints(t, "default", []string{"checkpoint_1.sql"})
		// Add a non-checkpoint file
		profileDir := filepath.Join(baseDir, "default")
		os.WriteFile(filepath.Join(profileDir, "notes.txt"), []byte("hi"), 0644)

		got, err := ListCheckpointFilenames(baseDir, "default")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0] != "checkpoint_1.sql" {
			t.Errorf("got %v, want [checkpoint_1.sql]", got)
		}
	})

	t.Run("creates profile dir if missing", func(t *testing.T) {
		baseDir := t.TempDir()

		got, err := ListCheckpointFilenames(baseDir, "newprofile")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("got %d files, want 0", len(got))
		}
		// Verify the dir was created
		if _, err := os.Stat(filepath.Join(baseDir, "newprofile")); os.IsNotExist(err) {
			t.Error("expected profile dir to be created")
		}
	})
}

func TestDeleteCheckpoint(t *testing.T) {
	t.Run("deletes existing checkpoint", func(t *testing.T) {
		baseDir := createTestCheckpoints(t, "default", []string{"checkpoint_1.sql", "checkpoint_2.sql"})
		profileDir := filepath.Join(baseDir, "default")

		_, err := DeleteCheckpoint(baseDir, "default", "checkpoint_1.sql", NamingModeSequential)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := os.Stat(filepath.Join(profileDir, "checkpoint_1.sql")); !os.IsNotExist(err) {
			t.Error("expected checkpoint_1.sql to be deleted")
		}
		if _, err := os.Stat(filepath.Join(profileDir, "checkpoint_2.sql")); err != nil {
			t.Error("expected checkpoint_2.sql to still exist")
		}
	})

	t.Run("deletes by short name", func(t *testing.T) {
		baseDir := createTestCheckpoints(t, "default", []string{"checkpoint_1_my-backup.sql", "checkpoint_2.sql"})
		profileDir := filepath.Join(baseDir, "default")

		filename, err := DeleteCheckpoint(baseDir, "default", "my-backup", NamingModeSequential)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filename != "checkpoint_1_my-backup.sql" {
			t.Errorf("got %s, want checkpoint_1_my-backup.sql", filename)
		}
		if _, err := os.Stat(filepath.Join(profileDir, "checkpoint_1_my-backup.sql")); !os.IsNotExist(err) {
			t.Error("expected checkpoint_1_my-backup.sql to be deleted")
		}
	})

	t.Run("returns error for non-existent checkpoint", func(t *testing.T) {
		baseDir := createTestCheckpoints(t, "default", nil)

		_, err := DeleteCheckpoint(baseDir, "default", "checkpoint_99.sql", NamingModeSequential)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestRenameCheckpoint(t *testing.T) {
	t.Run("rename a named sequential checkpoint", func(t *testing.T) {
		baseDir := createTestCheckpoints(t, "default", []string{"checkpoint_3_old-name.sql"})
		profileDir := filepath.Join(baseDir, "default")

		newFilename, err := RenameCheckpoint(baseDir, "default", "checkpoint_3_old-name.sql", "new-name", NamingModeSequential)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if newFilename != "checkpoint_3_new-name.sql" {
			t.Errorf("got %s, want checkpoint_3_new-name.sql", newFilename)
		}
		if _, err := os.Stat(filepath.Join(profileDir, "checkpoint_3_old-name.sql")); !os.IsNotExist(err) {
			t.Error("expected old file to be gone")
		}
		if _, err := os.Stat(filepath.Join(profileDir, "checkpoint_3_new-name.sql")); err != nil {
			t.Error("expected new file to exist")
		}
	})

	t.Run("add a name to an unnamed checkpoint", func(t *testing.T) {
		baseDir := createTestCheckpoints(t, "default", []string{"checkpoint_5.sql"})

		newFilename, err := RenameCheckpoint(baseDir, "default", "checkpoint_5.sql", "added-name", NamingModeSequential)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if newFilename != "checkpoint_5_added-name.sql" {
			t.Errorf("got %s, want checkpoint_5_added-name.sql", newFilename)
		}
	})

	t.Run("remove a name with empty string", func(t *testing.T) {
		baseDir := createTestCheckpoints(t, "default", []string{"checkpoint_2_some-name.sql"})

		newFilename, err := RenameCheckpoint(baseDir, "default", "checkpoint_2_some-name.sql", "", NamingModeSequential)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if newFilename != "checkpoint_2.sql" {
			t.Errorf("got %s, want checkpoint_2.sql", newFilename)
		}
	})

	t.Run("rename a named timestamp checkpoint", func(t *testing.T) {
		baseDir := createTestCheckpoints(t, "default", []string{"checkpoint_2026-03-02_15-30-45_old.sql"})

		newFilename, err := RenameCheckpoint(baseDir, "default", "checkpoint_2026-03-02_15-30-45_old.sql", "new", NamingModeTimestamp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if newFilename != "checkpoint_2026-03-02_15-30-45_new.sql" {
			t.Errorf("got %s, want checkpoint_2026-03-02_15-30-45_new.sql", newFilename)
		}
	})

	t.Run("rename a named unix checkpoint", func(t *testing.T) {
		baseDir := createTestCheckpoints(t, "default", []string{"checkpoint_1740934245_old.sql"})

		newFilename, err := RenameCheckpoint(baseDir, "default", "checkpoint_1740934245_old.sql", "new", NamingModeUnix)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if newFilename != "checkpoint_1740934245_new.sql" {
			t.Errorf("got %s, want checkpoint_1740934245_new.sql", newFilename)
		}
	})

	t.Run("rename by short name", func(t *testing.T) {
		baseDir := createTestCheckpoints(t, "default", []string{"checkpoint_3_old-name.sql"})
		profileDir := filepath.Join(baseDir, "default")

		newFilename, err := RenameCheckpoint(baseDir, "default", "old-name", "new-name", NamingModeSequential)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if newFilename != "checkpoint_3_new-name.sql" {
			t.Errorf("got %s, want checkpoint_3_new-name.sql", newFilename)
		}
		if _, err := os.Stat(filepath.Join(profileDir, "checkpoint_3_new-name.sql")); err != nil {
			t.Error("expected new file to exist")
		}
	})

	t.Run("error on non-existent target", func(t *testing.T) {
		baseDir := createTestCheckpoints(t, "default", nil)

		_, err := RenameCheckpoint(baseDir, "default", "checkpoint_99.sql", "name", NamingModeSequential)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestPruneCheckpoints(t *testing.T) {
	t.Run("prunes all but latest", func(t *testing.T) {
		files := []string{"checkpoint_1.sql", "checkpoint_2.sql", "checkpoint_3.sql"}
		baseDir := createTestCheckpoints(t, "default", files)
		profileDir := filepath.Join(baseDir, "default")

		count, err := PruneCheckpoints(baseDir, "default", NamingModeSequential)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 2 {
			t.Errorf("got count %d, want 2", count)
		}

		// Only checkpoint_3.sql should remain
		remaining, _ := os.ReadDir(profileDir)
		if len(remaining) != 1 {
			t.Fatalf("got %d remaining files, want 1", len(remaining))
		}
		if remaining[0].Name() != "checkpoint_3.sql" {
			t.Errorf("got %s, want checkpoint_3.sql", remaining[0].Name())
		}
	})

	t.Run("prunes named sequential checkpoints", func(t *testing.T) {
		files := []string{
			"checkpoint_1_init.sql",
			"checkpoint_2_before-migration.sql",
			"checkpoint_3_after-migration.sql",
		}
		baseDir := createTestCheckpoints(t, "default", files)
		profileDir := filepath.Join(baseDir, "default")

		count, err := PruneCheckpoints(baseDir, "default", NamingModeSequential)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 2 {
			t.Errorf("got count %d, want 2", count)
		}

		remaining, _ := os.ReadDir(profileDir)
		if len(remaining) != 1 {
			t.Fatalf("got %d remaining files, want 1", len(remaining))
		}
		if remaining[0].Name() != "checkpoint_3_after-migration.sql" {
			t.Errorf("got %s, want checkpoint_3_after-migration.sql", remaining[0].Name())
		}
	})

	t.Run("no-op with single checkpoint", func(t *testing.T) {
		baseDir := createTestCheckpoints(t, "default", []string{"checkpoint_1.sql"})

		count, err := PruneCheckpoints(baseDir, "default", NamingModeSequential)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 0 {
			t.Errorf("got count %d, want 0", count)
		}
	})

	t.Run("no-op with no checkpoints", func(t *testing.T) {
		baseDir := createTestCheckpoints(t, "default", nil)

		count, err := PruneCheckpoints(baseDir, "default", NamingModeSequential)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 0 {
			t.Errorf("got count %d, want 0", count)
		}
	})

	t.Run("prunes timestamp mode", func(t *testing.T) {
		files := []string{
			"checkpoint_2026-01-01_10-00-00.sql",
			"checkpoint_2026-02-15_08-00-00.sql",
			"checkpoint_2026-03-02_15-30-45.sql",
		}
		baseDir := createTestCheckpoints(t, "default", files)
		profileDir := filepath.Join(baseDir, "default")

		count, err := PruneCheckpoints(baseDir, "default", NamingModeTimestamp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 2 {
			t.Errorf("got count %d, want 2", count)
		}

		remaining, _ := os.ReadDir(profileDir)
		if len(remaining) != 1 {
			t.Fatalf("got %d remaining files, want 1", len(remaining))
		}
		if remaining[0].Name() != "checkpoint_2026-03-02_15-30-45.sql" {
			t.Errorf("got %s, want checkpoint_2026-03-02_15-30-45.sql", remaining[0].Name())
		}
	})

	t.Run("prunes named timestamp checkpoints", func(t *testing.T) {
		files := []string{
			"checkpoint_2026-01-01_10-00-00_init.sql",
			"checkpoint_2026-02-15_08-00-00_pre-deploy.sql",
			"checkpoint_2026-03-02_15-30-45_post-deploy.sql",
		}
		baseDir := createTestCheckpoints(t, "default", files)
		profileDir := filepath.Join(baseDir, "default")

		count, err := PruneCheckpoints(baseDir, "default", NamingModeTimestamp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 2 {
			t.Errorf("got count %d, want 2", count)
		}

		remaining, _ := os.ReadDir(profileDir)
		if len(remaining) != 1 {
			t.Fatalf("got %d remaining files, want 1", len(remaining))
		}
		if remaining[0].Name() != "checkpoint_2026-03-02_15-30-45_post-deploy.sql" {
			t.Errorf("got %s, want checkpoint_2026-03-02_15-30-45_post-deploy.sql", remaining[0].Name())
		}
	})

	t.Run("prunes named unix checkpoints", func(t *testing.T) {
		files := []string{
			"checkpoint_1740000000_init.sql",
			"checkpoint_1740934245_fresh-seed.sql",
		}
		baseDir := createTestCheckpoints(t, "default", files)
		profileDir := filepath.Join(baseDir, "default")

		count, err := PruneCheckpoints(baseDir, "default", NamingModeUnix)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 1 {
			t.Errorf("got count %d, want 1", count)
		}

		remaining, _ := os.ReadDir(profileDir)
		if len(remaining) != 1 {
			t.Fatalf("got %d remaining files, want 1", len(remaining))
		}
		if remaining[0].Name() != "checkpoint_1740934245_fresh-seed.sql" {
			t.Errorf("got %s, want checkpoint_1740934245_fresh-seed.sql", remaining[0].Name())
		}
	})
}
