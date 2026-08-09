#!/usr/bin/env python3
"""
Release Helper Script for Freezo

Analyzes git history, calculates SemVer version bumps, parses Helm Chart.yaml,
generates changelogs, and assists in executing release steps.
"""

import sys
import os
import re
import json
import subprocess
from pathlib import Path

def run_cmd(cmd, cwd=None):
    result = subprocess.run(cmd, shell=True, text=True, capture_output=True, cwd=cwd)
    if result.returncode != 0:
        raise Exception(f"Command failed ({cmd}): {result.stderr.strip()}")
    return result.stdout.strip()

def get_repo_root():
    return Path(run_cmd("git rev-parse --show-toplevel"))

def get_latest_tag():
    try:
        tags = run_cmd("git tag -l --sort=-v:refname").splitlines()
        semver_tags = [t for t in tags if re.match(r"^v?\d+\.\d+\.\d+$", t)]
        if semver_tags:
            return semver_tags[0]
    except Exception:
        pass
    return None

def parse_version(tag):
    if not tag:
        return (0, 0, 0)
    cleaned = tag.lstrip('v')
    parts = cleaned.split('.')
    return (int(parts[0]), int(parts[1]), int(parts[2]))

def get_commits_since_tag(tag):
    if tag:
        rev_range = f"{tag}..HEAD"
    else:
        rev_range = "HEAD"
    
    lines = run_cmd(f'git log {rev_range} --pretty=format:"%h%x09%s%x09%b%x1f"').split('\x1f')
    commits = []
    for line in lines:
        line = line.strip()
        if not line:
            continue
        parts = line.split('\x09')
        hash_val = parts[0].strip()
        subject = parts[1].strip() if len(parts) > 1 else ""
        body = parts[2].strip() if len(parts) > 2 else ""
        commits.append({
            "hash": hash_val,
            "subject": subject,
            "body": body
        })
    return commits

def determine_bump(commits):
    has_breaking = False
    has_feat = False
    has_patch = False

    for c in commits:
        subj = c["subject"]
        body = c["body"]
        if "BREAKING CHANGE:" in body or "BREAKING CHANGE:" in subj or re.match(r"^\w+(\([\w\-\.]+\))?!:", subj):
            has_breaking = True
        elif subj.startswith("feat") or re.match(r"^feat(\([\w\-\.]+\))?:", subj):
            has_feat = True
        else:
            has_patch = True

    if has_breaking:
        return "major"
    elif has_feat:
        return "minor"
    elif has_patch or len(commits) > 0:
        return "patch"
    else:
        return "none"

def bump_version(current_tuple, bump_type):
    major, minor, patch = current_tuple
    if bump_type == "major":
        return f"v{major + 1}.0.0"
    elif bump_type == "minor":
        return f"v{major}.{minor + 1}.0"
    elif bump_type == "patch":
        return f"v{major}.{minor}.{patch + 1}"
    else:
        return f"v{major}.{minor}.{patch}"

def generate_changelog(commits):
    features = []
    fixes = []
    maintenance = []
    breaking = []

    for c in commits:
        subj = c["subject"]
        body = c["body"]
        h = c["hash"]
        item = f"- {subj} (`{h}`)"

        if "BREAKING CHANGE:" in body or "BREAKING CHANGE:" in subj or "!" in subj.split(":")[0]:
            breaking.append(item)
        
        if subj.startswith("feat") or re.match(r"^feat(\([\w\-\.]+\))?:", subj):
            features.append(item)
        elif subj.startswith("fix") or re.match(r"^fix(\([\w\-\.]+\))?:", subj):
            fixes.append(item)
        else:
            maintenance.append(item)

    sections = []
    if breaking:
        sections.append("### ⚠️ Breaking Changes\n" + "\n".join(breaking))
    if features:
        sections.append("### 🚀 Features\n" + "\n".join(features))
    if fixes:
        sections.append("### 🐛 Bug Fixes\n" + "\n".join(fixes))
    if maintenance:
        sections.append("### 🧰 Maintenance & Refactoring\n" + "\n".join(maintenance))

    if not sections:
        return "No notable changes."
    return "\n\n".join(sections)

def get_helm_chart_info(repo_root):
    chart_path = repo_root / "charts" / "freezo" / "Chart.yaml"
    if not chart_path.exists():
        return None, None, chart_path
    
    content = chart_path.read_text()
    version_match = re.search(r"^version:\s*(.+)$", content, re.MULTILINE)
    app_version_match = re.search(r"^appVersion:\s*(.+)$", content, re.MULTILINE)

    chart_version = version_match.group(1).strip('"\'') if version_match else None
    app_version = app_version_match.group(1).strip('"\'') if app_version_match else None
    return chart_version, app_version, chart_path

def bump_chart_version(chart_version):
    if not chart_version:
        return "0.1.0"
    parts = chart_version.split('.')
    if len(parts) == 3 and parts[2].isdigit():
        return f"{parts[0]}.{parts[1]}.{int(parts[2]) + 1}"
    return chart_version

def analyze():
    repo_root = get_repo_root()
    latest_tag = get_latest_tag()
    current_tuple = parse_version(latest_tag)
    current_version_str = f"v{current_tuple[0]}.{current_tuple[1]}.{current_tuple[2]}" if latest_tag else "v0.0.0"

    commits = get_commits_since_tag(latest_tag)
    bump_type = determine_bump(commits)
    new_version = bump_version(current_tuple, bump_type)

    chart_version, app_version, chart_path = get_helm_chart_info(repo_root)
    new_chart_version = bump_chart_version(chart_version)

    changelog = generate_changelog(commits)

    return {
        "repo_root": str(repo_root),
        "latest_tag": latest_tag or "None",
        "current_version": current_version_str,
        "bump_type": bump_type,
        "new_version": new_version,
        "commit_count": len(commits),
        "chart_path": str(chart_path),
        "current_chart_version": chart_version,
        "new_chart_version": new_chart_version,
        "current_app_version": app_version,
        "new_app_version": new_version,
        "changelog": changelog
    }

def main():
    if len(sys.argv) > 1 and sys.argv[1] == "apply":
        # Apply helm chart changes
        info = analyze()
        chart_path = Path(info["chart_path"])
        if chart_path.exists():
            content = chart_path.read_text()
            content = re.sub(r"^version:\s*.+$", f"version: {info['new_chart_version']}", content, flags=re.MULTILINE)
            content = re.sub(r"^appVersion:\s*.+$", f'appVersion: "{info["new_app_version"]}"', content, flags=re.MULTILINE)
            chart_path.write_text(content)
            print(f"Updated {chart_path} with version: {info['new_chart_version']} and appVersion: {info['new_app_version']}")
    else:
        info = analyze()
        print(json.dumps(info, indent=2))

if __name__ == "__main__":
    main()
