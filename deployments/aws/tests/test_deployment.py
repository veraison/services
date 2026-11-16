# Copyright 2026 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
"""Offline deployment regressions; AWS clients and build commands are mocked."""
import argparse
import ast
from collections import defaultdict
import logging
import os
from pathlib import Path
import re
import shlex
import shutil
import string
import subprocess
import sys
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import Mock, patch
from urllib.parse import urlparse

import yaml


ROOT = Path(__file__).resolve().parents[3]
AWS = ROOT / "deployments/aws"
TREE = ast.parse((AWS / "bin/veraison").read_text())


class ClientError(Exception):
    def __init__(self, code, message):
        self.response = {"Error": {"Code": code, "Message": message}}
        super().__init__(message)


def definitions(*names, **extra):
    """Load the real definitions without importing optional deployment clients."""
    namespace = dict(
        argparse=argparse, defaultdict=defaultdict, logging=logging, os=os,
        re=re, shlex=shlex, shutil=shutil, string=string, tempfile=tempfile,
        ClientError=ClientError, urlparse=urlparse, time=Mock(),
        __file__=str(AWS / "bin/veraison"),
    )
    namespace.update(extra)
    nodes = [
        node for node in TREE.body
        if isinstance(node, (ast.ClassDef, ast.FunctionDef))
        and node.name in names
    ]
    assert len(nodes) == len(names)
    exec(compile(ast.Module(body=nodes, type_ignores=[]), str(AWS / "bin/veraison"), "exec"),
         namespace)
    return namespace


class CliTests(unittest.TestCase):
    def command(self, name, **extra):
        ns = definitions("BaseCommand", name, **extra)
        cmd = ns[name](Mock())
        cmd.cache = {}
        cmd.logger = Mock()
        return cmd, ns

    def test_sdk_receives_deployment_region_without_aws_default_region(self):
        main = next(node for node in TREE.body if isinstance(node, ast.If)
                    and ast.unparse(node.test) == "__name__ == '__main__'")
        assignment = next(node for node in main.body if isinstance(node, ast.Assign)
                          and any(isinstance(t, ast.Name) and t.id == "aws"
                                  for t in node.targets))
        code = compile(ast.Module(body=[assignment], type_ignores=[]), "<AWS setup>", "exec")
        factory = Mock()
        with patch.dict(os.environ, {"VERAISON_AWS_REGION": "us-east-1"}, clear=True):
            exec(code, {"Aws": factory, "os": os})
        self.assertEqual(factory.call_args.kwargs["region_name"], "us-east-1")

    def test_package_path_uses_marker_instead_of_dpkg_language_or_package_name(self):
        with tempfile.TemporaryDirectory(prefix="veraison test ") as work:
            root = Path(work)
            script = root / "src/deployments/debian/deployment.sh"
            script.parent.mkdir(parents=True)
            script.touch()
            deb = root / "veraison_1.0_amd64.deb"
            deb.write_bytes(b"package")
            cache = root / "cache"
            cache.mkdir()
            build = Mock(return_value=(0, f"localized dpkg output\nVERAISON_DEB={deb}\ndone.\n", ""))
            cmd, _ = self.command("CreateDebCommand", run_in_shell=build)
            args = SimpleNamespace(veraison_src=str(root / "src"),
                                   work_dir=str(root), cache_dir=str(cache), verbose=False)
            with patch.dict(os.environ):
                cmd.run(args)
            self.assertEqual(Path(cmd.cache["deb"]).read_bytes(), b"package")
            self.assertEqual(shlex.split(build.call_args.args[0]),
                             [str(script), "create-deb", str(root)])

    def test_failed_sentinel_tool_build_never_starts_packer(self):
        create_image = Mock()
        cmd, _ = self.command(
            "CreateSentinelImageCommand", run_in_shell=Mock(return_value=(1, "", "Go failed")),
            command_create_image=create_image)
        cmd.fail_shell = Mock(side_effect=RuntimeError("tool build failed"))
        with self.assertRaisesRegex(RuntimeError, "tool build failed"):
            cmd.run(SimpleNamespace(package="example/tool", rev="123", verbose=False))
        create_image.assert_not_called()

    def test_sentinel_tool_revision_matches_services_dependency(self):
        cmd, _ = self.command("CreateSentinelImageCommand")
        parser = argparse.ArgumentParser()
        cmd.update_arguments(parser)
        revision = parser.parse_args([]).rev
        version = re.search(r"github.com/veraison/corim-store\s+(\S+)",
                            (ROOT / "go.mod").read_text()).group(1)
        self.assertTrue(version.endswith("-" + revision))

    def test_cached_signer_is_reused_without_generating_or_replacing_key(self):
        jwk = Mock()
        cmd, _ = self.command("CreateSignerKeyCommand", jwk=jwk)
        cmd.cache["signer-key-arn"] = "arn:example:signer"
        cmd.aws.secretsmanager.describe_secret.return_value = {"Name": "demo-ear-signer"}
        cmd.run(SimpleNamespace(deployment_name="demo"))
        jwk.JWK.generate.assert_not_called()
        cmd.aws.secretsmanager.create_secret.assert_not_called()

    def test_deleted_cached_signer_is_rejected(self):
        cmd, _ = self.command("CreateSignerKeyCommand", jwk=Mock())
        cmd.cache["signer-key-arn"] = "arn:example:signer"
        cmd.aws.secretsmanager.describe_secret.return_value = {
            "Name": "demo-ear-signer", "DeletedDate": "pending deletion"}
        with self.assertRaises(RuntimeError):
            cmd.run(SimpleNamespace(deployment_name="demo"))
        cmd.aws.secretsmanager.create_secret.assert_not_called()

    def test_command_exception_has_nonzero_exit_unless_no_error_requested(self):
        main = next(node for node in TREE.body if isinstance(node, ast.If)
                    and ast.unparse(node.test) == "__name__ == '__main__'")
        handler = next(node for node in reversed(main.body) if isinstance(node, ast.Try))
        code = compile(ast.Module(body=[handler], type_ignores=[]), "<CLI dispatch>", "exec")
        for fail, expected in ((True, 1), (False, 0)):
            with self.subTest(fail_with_error=fail):
                cmd = Mock()
                cmd.execute.side_effect = RuntimeError("failed")
                with self.assertRaises(SystemExit) as caught:
                    exec(code, {"cmd": cmd, "args": SimpleNamespace(fail_with_error=fail),
                                "sys": sys})
                self.assertEqual(caught.exception.code, expected)

    def test_deleted_failed_stack_reports_original_creation_failure(self):
        ns = definitions("command_create_stack", camel_to_kebab=lambda key: key)
        cmd = Mock()
        cmd.fail.side_effect = RuntimeError("could not create stack support")
        cmd.aws.cf.create_stack.return_value = {"StackId": "example"}
        cmd.aws.cf.describe_stacks.side_effect = [
            {"Stacks": [{"StackStatus": "DELETE_IN_PROGRESS"}]},
            ClientError("ValidationError", "Stack with id demo-support does not exist"),
        ]
        cmd.aws.cf.describe_stack_events.return_value = {"StackEvents": [{
            "ResourceStatus": "CREATE_FAILED", "ResourceStatusReason": "invalid node type"}]}
        with tempfile.NamedTemporaryFile(mode="w") as template:
            template.write("{}")
            template.flush()
            with self.assertRaisesRegex(RuntimeError, "could not create stack support"):
                ns["command_create_stack"](cmd, "demo", "support", template.name, [], 0)
        cmd.logger.error.assert_any_call("CREATE_FAILED invalid node type")


class DiscoveryTests(unittest.TestCase):
    def setUp(self):
        self.requests = SimpleNamespace(RequestException=OSError)
        self.read = definitions("command_get_discovery",
                                requests=self.requests)["command_get_discovery"]
        self.cmd = Mock()
        self.cmd.fail.side_effect = RuntimeError("discovery failed")
        self.session = Mock()

    def test_success_uses_timeout_and_checks_http_status(self):
        self.session.get.return_value.json.return_value = {"version": "1"}
        self.assertEqual(self.read(self.cmd, "https://example.test", ("version",),
                                   self.session), {"version": "1"})
        self.session.get.assert_called_once_with("https://example.test", timeout=(10, 30))
        self.session.get.return_value.raise_for_status.assert_called_once()

    def test_http_error_does_not_turn_into_missing_version(self):
        self.session.get.return_value.raise_for_status.side_effect = OSError("503")
        with self.assertRaises(RuntimeError):
            self.read(self.cmd, "https://example.test", ("version",), self.session)
        self.session.get.return_value.json.assert_not_called()
        self.assertIn("503", self.cmd.fail.call_args.args[0])

    def test_non_json_response_is_reported(self):
        self.session.get.return_value.json.side_effect = ValueError("not JSON")
        with self.assertRaises(RuntimeError):
            self.read(self.cmd, "https://example.test", ("version",), self.session)

    def test_missing_fields_and_non_object_json_are_reported(self):
        for document in ({"error": "not ready"}, []):
            with self.subTest(document=document):
                self.session.get.return_value.json.return_value = document
                with self.assertRaises(RuntimeError):
                    self.read(self.cmd, "https://example.test", ("version",), self.session)


class ConfigurationTests(unittest.TestCase):
    def test_make_bootstrap_keeps_the_sourced_environment(self):
        with tempfile.TemporaryDirectory() as work:
            root = Path(work)
            shutil.copy2(AWS / "Makefile", root / "Makefile")
            (root / "env").mkdir()
            (root / "env/env.bash").write_text("export BOOTSTRAP_MARKER=present\n")
            script = root / "deployment.sh"
            script.write_text('#!/bin/bash\n[[ "$BOOTSTRAP_MARKER" == present ]]\n')
            script.chmod(0o755)
            subprocess.run(["make", "-C", str(root), "bootstrap"], check=True,
                           capture_output=True, text=True, timeout=20)

    def test_template_expansion_retains_cloudwatch_placeholders(self):
        render = definitions("command_instantiate_template")["command_instantiate_template"]
        with tempfile.TemporaryDirectory() as work:
            source = Path(work) / "sample.template"
            source.write_text("$REGION $UNSET $@{aws:InstanceId}")
            with patch.dict(os.environ, {"REGION": "test-region"}, clear=True):
                output = render(Mock(), "demo", str(source), work)
            self.assertEqual(Path(output).read_text(),
                             "test-region  " + "$" + "{aws:InstanceId}")

    def test_services_use_current_plugins_and_encrypted_stores(self):
        config = yaml.safe_load((AWS / "templates/services-config.yaml.template").read_text())
        plugin = config["plugin"]
        self.assertEqual(plugin["go-plugin-schemes"]["dir"], "/opt/veraison/plugins/schemes/")
        self.assertEqual(plugin["go-plugin-stores"]["dir"], "/opt/veraison/plugins/stores/")
        store = config["endorsement-store"]
        self.assertEqual(store["active-plugins"], ["corim-store"])
        self.assertEqual(store["plugin-parameters"]["corim-store"]["dbms"], "pgx")
        self.assertIn("sslmode=require", store["plugin-parameters"]["corim-store"]["dsn"])
        self.assertIn("sslmode=require", config["po-store"]["sql"]["datasource"])
        self.assertNotIn("en-store", config)
        self.assertNotIn("ta-store", config)

    def test_env_sources_config_under_nounset_and_exports_region(self):
        with tempfile.TemporaryDirectory() as work:
            cfg = Path(work) / "account.cfg"
            cfg.write_text("VERAISON_AWS_REGION=us-east-1\n")
            env = {**os.environ, "VERAISON_AWS_VENV": str(Path(work) / "absent")}
            result = subprocess.run(
                ["bash", "-uc", 'source "$1" "$2"; printf "%s %s" "$AWS_DEFAULT_REGION" "$AWS_REGION"',
                 "--", str(AWS / "env/env.bash"), str(cfg)],
                env=env, text=True, capture_output=True, check=True, timeout=20)
            self.assertEqual(result.stdout, "us-east-1 us-east-1")
            subprocess.run(["bash", "-uc", 'env_file=$1; set --; source "$env_file"', "--", str(AWS / "env/env.bash")],
                           env=env, check=True, capture_output=True, timeout=20)

    def test_bringup_builds_for_unset_or_empty_deb_and_adds_explicit_deb(self):
        with tempfile.TemporaryDirectory() as work:
            root = Path(work)
            (root / "bin").mkdir()
            for file in ("deployment.sh", "deployment.cfg"):
                shutil.copy2(AWS / file, root / file)
            cli = root / "bin/veraison"
            cli.write_text('#!/bin/bash\nprintf "%s\\n" "$*" >> "$CALL_LOG"\n'
                           'if [[ "$*" == *create-deb && "$FAIL_PACKAGE" == yes ]]; then exit 23; fi\n'
                           'exit 0\n')
            cli.chmod(0o755)
            packer = root / "bin/packer"
            packer.write_text("#!/bin/bash\nexit 0\n")
            packer.chmod(0o755)
            env = {**os.environ, "PATH": str(root / "bin") + os.pathsep + os.environ["PATH"],
                   "VERAISON_AWS_REGION": "us-east-1", "VERAISON_AWS_ADMIN_CIDR": "192.0.2.1/32",
                   "VERAISON_AWS_VPC_CIDR": "10.1.0.0/16",
                   "VERAISON_AWS_DNS_NAME": "example.test", "FAIL_PACKAGE": "no"}
            for deb in (None, "", "/tmp/example.deb"):
                with self.subTest(deb=deb):
                    log = root / "calls"
                    log.write_text("")
                    env["CALL_LOG"] = str(log)
                    if deb is None:
                        env.pop("VERAISON_DEB", None)
                    else:
                        env["VERAISON_DEB"] = deb
                    subprocess.run(["bash", str(root / "deployment.sh"), "bringup"], env=env,
                                   check=True, text=True, capture_output=True, timeout=20)
                    calls = log.read_text()
                    self.assertIn("add-deb /tmp/example.deb" if deb else "create-deb", calls)
            env.pop("VERAISON_DEB", None)
            env["FAIL_PACKAGE"] = "yes"
            log.write_text("")
            result = subprocess.run(["bash", str(root / "deployment.sh"), "bringup"],
                                    env=env, capture_output=True, text=True, timeout=20)
            self.assertEqual(result.returncode, 23)
            self.assertNotIn("create-key-pair", log.read_text())


class EndToEndTests(unittest.TestCase):
    def test_provision_failure_stops_before_second_bundle(self):
        with tempfile.TemporaryDirectory() as work:
            root = Path(work)
            end_to_end = root / "end-to-end"
            end_to_end.mkdir()
            script = end_to_end / "end-to-end-aws"
            shutil.copy2(ROOT / "end-to-end/end-to-end-aws", script)
            env_dir = root / "deployments/aws/env"
            env_dir.mkdir(parents=True)
            (env_dir / "env.bash").write_text("")
            bin_dir = root / "bin"
            bin_dir.mkdir()
            for tool in ("veraison", "evcli", "cocli"):
                exe = bin_dir / tool
                exe.write_text('#!/bin/bash\n' + (
                    'echo submit >> "$CALL_LOG"\nexit "$COCLI_EXIT"\n'
                    if tool == "cocli" else "exit 0\n"))
                exe.chmod(0o755)
            for status, submissions in ((7, 1), (0, 2)):
                with self.subTest(status=status):
                    log = root / "calls"
                    log.write_text("")
                    env = {**os.environ, "PATH": str(bin_dir) + os.pathsep + os.environ["PATH"],
                           "SCHEME": "cca", "COCLI_EXIT": str(status), "CALL_LOG": str(log)}
                    result = subprocess.run(["bash", str(script), "provision"],
                                            env=env, text=True, capture_output=True, timeout=20)
                    self.assertEqual(result.returncode, status)
                    self.assertEqual(len(log.read_text().splitlines()), submissions)


if __name__ == "__main__":
    unittest.main()
