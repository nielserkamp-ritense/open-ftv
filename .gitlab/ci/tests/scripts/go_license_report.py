#!/usr/bin/env python3
"""
Checks the Go dependency tree's licenses against the EUPL compatibility allowlist, combining two detectors
to cover each other's gaps:

  1. cyclonedx-gomod scans each dependency's LICENSE/COPYING file and identifies the license from its text
     when it builds the SBOM. This script just reads that result.
  2. `licensecheck` scans source files for `// SPDX-License-Identifier:` comments. This script runs it only on
     dependencies cyclonedx-gomod could not classify.
"""
import argparse
import json
import re
import subprocess
import sys
from collections import Counter
from pathlib import Path

# EUPL v1.2 compatibility allowlist, shared with the .eupl_license_report jq script.
ALLOWED_LICENSES_FILE = Path(__file__).parent / "allowed_licenses.json"
ALLOWED_LICENSES = set(json.loads(ALLOWED_LICENSES_FILE.read_text()))
ROOT_LICENSE_FILENAMES = [
    "LICENSE", "LICENSE.txt", "LICENSE.md",
    "COPYING", "COPYING.txt", "COPYING.md",
    "UNLICENSE",
]
FIXTURE_DIR_NAMES = {"testdata", "test_files", "tests", "fixtures", "fixture"}
MAJORITY_THRESHOLD = 0.5


def license_tokens(expr: str) -> list[str]:
    return [
        t.strip()
        for t in re.split(r"[()]|,|\s+(?:and/or|OR|AND)\s+", expr)
        if t.strip()
    ]


def is_compliant(expr: str | None) -> bool:
    if not expr or expr == "UNKNOWN":
        return False
    tokens = license_tokens(expr)
    return bool(tokens) and all(t in ALLOWED_LICENSES for t in tokens)


def sbom_component_license(component: dict) -> str | None:
    licenses = component.get("licenses") or []
    parts = []
    for entry in licenses:
        lic = entry.get("license", {})
        parts.append(lic.get("id") or lic.get("name") or entry.get("expression"))
    parts = [p for p in parts if p]
    return ", ".join(parts) if parts else None


def resolve_module_dirs(name_versions: list[str]) -> dict[str, Path]:
    """Resolve each "name@version" string to its on-disk module directory via `go list -m`, batched in one call."""
    if not name_versions:
        return {}
    proc = subprocess.run(
        ["go", "list", "-m", "-f", "{{.Path}}@{{.Version}} {{.Dir}}", *name_versions],
        capture_output=True, text=True,
    )
    dirs = {}
    for line in proc.stdout.splitlines():
        if " " not in line:
            continue
        name_version, module_dir = line.split(" ", 1)
        if module_dir and module_dir != "<nil>":
            dirs[name_version] = Path(module_dir)
    return dirs


def run_licensecheck(licensecheck_bin: str, targets: list[Path]) -> dict[str, str]:
    if not targets:
        return {}
    cmd = [
        licensecheck_bin, "-r", "--machine", "--shortname-scheme=spdx",
        *[str(t) for t in targets],
    ]
    proc = subprocess.run(cmd, capture_output=True, text=True)
    results = {}
    for line in proc.stdout.splitlines():
        if "\t" not in line:
            continue
        file_path, license_expr = line.split("\t", 1)
        results[file_path] = license_expr.strip()
    return results


def has_fixture_dir(rel_path_within_module: str) -> bool:
    parts = rel_path_within_module.split("/")[:-1]
    return any(p in FIXTURE_DIR_NAMES for p in parts)


def find_root_license_candidates(module_dir: Path) -> dict[str, Path]:
    if not module_dir.is_dir():
        return {}
    by_lower = {p.name.lower(): p for p in module_dir.iterdir() if p.is_file()}
    found = {}
    for canonical in ROOT_LICENSE_FILENAMES:
        match = by_lower.get(canonical.lower())
        if match is not None:
            found[canonical] = match
    return found


def resolve_root_license(candidates: dict[str, Path], results: dict[str, str]) -> str | None:
    for canonical in ROOT_LICENSE_FILENAMES:
        match = candidates.get(canonical)
        if match is None:
            continue
        expr = results.get(str(match))
        if expr and expr != "UNKNOWN":
            return expr
    return None


def resolve_majority_license(module_dir: Path, results: dict[str, str]) -> str | None:
    votes: Counter = Counter()
    total = 0
    prefix = str(module_dir) + "/"
    for file_path, expr in results.items():
        if not file_path.startswith(prefix) or not file_path.endswith(".go"):
            continue
        within = file_path[len(prefix):]
        if has_fixture_dir(within):
            continue
        total += 1
        if expr and expr != "UNKNOWN":
            votes[expr] += 1
    if not votes or total == 0:
        return None
    expr, count = votes.most_common(1)[0]
    if count >= total * MAJORITY_THRESHOLD:
        return expr
    return None


def licensecheck_fallback(name_versions: list[str], licensecheck_bin: str) -> dict[str, str | None]:
    module_dirs = resolve_module_dirs(name_versions)

    candidates_by_module: dict[str, dict[str, Path]] = {}
    root_files: list[Path] = []
    for name_version, module_dir in module_dirs.items():
        candidates = find_root_license_candidates(module_dir)
        candidates_by_module[name_version] = candidates
        root_files.extend(candidates.values())

    pass1_results = run_licensecheck(licensecheck_bin, root_files)

    resolved: dict[str, str | None] = {}
    unresolved: list[tuple[str, Path]] = []
    for name_version, module_dir in module_dirs.items():
        expr = resolve_root_license(candidates_by_module[name_version], pass1_results)
        if expr:
            resolved[name_version] = expr
        else:
            unresolved.append((name_version, module_dir))

    if unresolved:
        pass2_results = run_licensecheck(licensecheck_bin, [d for _, d in unresolved])
        for name_version, module_dir in unresolved:
            resolved[name_version] = resolve_majority_license(module_dir, pass2_results)

    return resolved


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--sbom", type=Path, required=True)
    parser.add_argument("--licensecheck-bin", default="licensecheck")
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()

    sbom = json.loads(args.sbom.read_text())
    components = sbom.get("components", [])

    flagged: dict[str, str | None] = {}
    for component in components:
        name = component.get("name")
        version = component.get("version")
        if not name or not version:
            continue
        expr = sbom_component_license(component)
        if not is_compliant(expr):
            flagged[f"{name}@{version}"] = expr

    fallback_results = licensecheck_fallback(list(flagged.keys()), args.licensecheck_bin)

    final_issues = []
    for name_version, original_expr in flagged.items():
        fallback_expr = fallback_results.get(name_version)
        # Prefer whichever detector actually found something; if both did,
        # prefer the fallback since it was specifically checked because the
        # first pass flagged this component.
        effective_expr = fallback_expr or original_expr
        if is_compliant(effective_expr):
            continue
        label = effective_expr if effective_expr else "UNKNOWN"
        final_issues.append(f"  - {name_version}: {label}")

    if final_issues:
        args.output.write_text("\n".join(sorted(final_issues)) + "\n")
        print("\033[1;33mWARNING: Go dependencies found with a license that is not EUPL-compatible (or could not be detected):\033[0m")
        print("\n".join(sorted(final_issues)))
    else:
        args.output.write_text("")
        print("All Go dependency licenses are EUPL-compatible.")

    return 0


if __name__ == "__main__":
    sys.exit(main())
