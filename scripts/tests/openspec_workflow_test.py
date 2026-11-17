"""Exercise the pinned real OpenSpec CLI and repository creation boundary.

Run explicitly after npm ci --prefix tools/openspec. No harness/source facts are
fabricated: these are isolated workflow/schema fixtures, not telemetry evidence.
"""
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
CLI = ROOT / 'scripts/openspec.sh'
CREATE = ROOT / 'scripts/openspec-change.mjs'


class OpenSpecWorkflowTest(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix='al84-openspec-test-')
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.project = self.root / 'selected project'
        self.project.mkdir()
        self.env = dict(os.environ, OPENSPEC_TELEMETRY='0', DO_NOT_TRACK='1', CI='1')
        for key in ('HOME', 'XDG_CONFIG_HOME', 'XDG_CACHE_HOME', 'XDG_DATA_HOME'):
            target = self.root / key.lower()
            target.mkdir()
            self.env[key] = str(target)
        self.project.joinpath('.agenticlens-config.yaml').write_text(
            'ticket:\n  prefixes:\n    - prefix: TASK-\n      description: Fixture work\n'
            'workspaces:\n  roots:\n    - path: openspec/changes\n'
            '      discovery:\n        max_depth: 1\n        max_directories: 300\n')
        shutil.copytree(ROOT / 'openspec/schemas', self.project / 'openspec/schemas')
        shutil.copyfile(ROOT / 'openspec/config.yaml', self.project / 'openspec/config.yaml')

    def run_cli(self, *args):
        result = subprocess.run([str(CLI), *args], cwd=self.project, env=self.env,
                                text=True, capture_output=True, timeout=60)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        return result

    def create(self, *args):
        return subprocess.run(['node', str(CREATE), *args], cwd=self.project,
                              env=self.env, text=True, capture_output=True, timeout=60)

    def test_creation_preserves_canonical_identity_and_rejects_collisions(self):
        result = self.create('TASK-42', 'selected implementation')
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertIn('Canonical ticket: TASK-42', result.stdout)
        change = self.project / 'openspec/changes/task-42-selected-implementation'
        self.assertTrue(change.joinpath('.openspec.yaml').is_file())
        self.assertFalse(change.joinpath('agentic_metrics.json').exists())
        self.assertFalse(change.joinpath('.agenticlens-workspace.json').exists())
        sentinel = change / 'custom.md'
        sentinel.write_text('preserve custom content\n')
        original = change.joinpath('.openspec.yaml').read_bytes()
        duplicate = self.create('TASK-42', 'selected implementation')
        self.assertNotEqual(duplicate.returncode, 0)
        self.assertIn('refusing to overwrite', duplicate.stderr)
        self.assertEqual(sentinel.read_text(), 'preserve custom content\n')
        self.assertEqual(change.joinpath('.openspec.yaml').read_bytes(), original)
        for invalid in ('OTHER-42', 'task-42', 'TASK-../42'):
            rejected = self.create(invalid, 'rejected')
            self.assertNotEqual(rejected.returncode, 0, invalid)
        self.assertEqual(sorted(p.name for p in change.parent.iterdir() if p.is_dir()),
                         ['archive', 'task-42-selected-implementation'])

    def test_optional_setup_bundle_preserves_custom_files_before_any_install(self):
        installer = ROOT / '.agents/skills/agentic-lens-setup/scripts/install.sh'
        target = self.root / 'adopting repository'
        target.mkdir()
        command = ['bash', str(installer), '--version', 'v9.8.7', '--source-dir',
                   str(ROOT), '--dest', str(target), '--with-openspec']
        config = target / 'openspec/config.yaml'
        config.parent.mkdir()
        config.write_text('schema: custom-owner-schema\n')
        rejected = subprocess.run(command, env=self.env, text=True, capture_output=True, timeout=30)
        self.assertNotEqual(rejected.returncode, 0)
        self.assertIn('Existing optional OpenSpec file differs', rejected.stderr)
        self.assertEqual(config.read_text(), 'schema: custom-owner-schema\n')
        self.assertFalse((target / '.agents').exists(), 'Preflight must precede all bundle writes')
        config.unlink()
        result = subprocess.run(command, env=self.env, text=True, capture_output=True, timeout=30)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        self.assertTrue((target / 'tools/openspec/package-lock.json').is_file())
        self.assertTrue((target / 'openspec/schemas/agenticlens-planning/schema.yaml').is_file())
        self.assertTrue((target / 'scripts/openspec-change.mjs').is_file())
        self.assertFalse((target / 'tools/openspec/node_modules').exists())
        self.assertFalse((target / 'openspec/changes').exists(), 'Installing a bundle must not create execution work')
        self.assertFalse((target / '.agenticlens-config.local.yaml').exists())
        repeated = subprocess.run(command, env=self.env, text=True, capture_output=True, timeout=30)
        self.assertEqual(repeated.returncode, 0, repeated.stdout + repeated.stderr)

    def test_planning_schema_has_no_developer_contract_and_archives_evidence(self):
        self.run_cli('schema', 'validate', 'agenticlens', '--json')
        self.run_cli('schema', 'validate', 'agenticlens-planning', '--json')
        result = self.create('TASK-43', 'planning delivery', '--planning')
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        change = self.project / 'openspec/changes/task-43-planning-delivery'
        metadata = change.joinpath('.openspec.yaml').read_text()
        self.assertIn('schema: agenticlens-planning', metadata)
        self.assertIn('skip_specs: true', metadata)
        status = json.loads(self.run_cli('status', '--change', change.name, '--json').stdout)
        self.assertEqual([item['id'] for item in status['artifacts']],
                         ['proposal', 'tasks', 'delivery'])
        self.assertFalse(change.joinpath('design.md').exists())
        change.joinpath('proposal.md').write_text('# Planning fixture\n\nRoadmap outputs only.\n')
        change.joinpath('tasks.md').write_text('# Tasks\n\n- [x] Verify planning schema.\n')
        change.joinpath('delivery.md').write_text('# docs(TASK-43): planning fixture\n\nNo implementation authority.\n')
        self.run_cli('validate', change.name, '--strict')
        apply = json.loads(self.run_cli('instructions', 'apply', '--change', change.name, '--json').stdout)
        self.assertEqual(apply['state'], 'all_done')
        self.assertNotIn('design', apply['contextFiles'])
        self.assertNotIn('specs', apply['contextFiles'])
        self.run_cli('archive', change.name, '--yes', '--json')
        archived = list((change.parent / 'archive').glob('*-task-43-planning-delivery'))
        self.assertEqual(len(archived), 1)
        self.assertIn('No implementation authority.', archived[0].joinpath('delivery.md').read_text())
        self.assertFalse(change.exists())


if __name__ == '__main__':
    unittest.main()
