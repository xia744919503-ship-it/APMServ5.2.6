from __future__ import annotations

import json
from pathlib import Path

ROOT = Path(r"D:\APMServ5.2.6\new_project")
PROBE = ROOT / "artifacts" / "swf-probe" / "BloodWar.json"
OUT = ROOT / "artifacts" / "swf-probe" / "target-classes.md"

TARGETS = (
    "Login.LoginDialog",
    "Login.CreateRoleDialog",
    "Login.JackarooDialog",
    "Building.BuildingGrid",
    "Building.CreateBuildingDialog",
    "Building.CreateBuildingItem",
    "Building.BuildingListDialog",
    "Building.BuildingTip",
    "Bar.TopPanel",
    "Bar.BottomPanel",
    "UserInfo.InfoPanel",
    "guide.GuideTip",
)

KIND = {
    0: "slot",
    1: "method",
    2: "getter",
    3: "setter",
    4: "class",
    5: "function",
    6: "const",
}


def main() -> None:
    data = json.loads(PROBE.read_text(encoding="utf-8"))
    instances = []
    methods = {}
    bodies = {}
    for block in data["abcBlocks"]:
        for method in block.get("abcSummary", {}).get("interestingMethods", []):
            methods[method["index"]] = method
        for body in block.get("abcSummary", {}).get("methodBodies", []):
            bodies[body["method"]] = body
        for item in block.get("abcSummary", {}).get("interestingInstances", []):
            if item["name"] in TARGETS:
                instances.append(item)

    lines = ["# Target ABC Classes", ""]
    for item in instances:
        lines.extend(
            [
                f"## {item['name']}",
                "",
                f"- Extends: `{item.get('super','')}`",
                f"- Init method: `{item.get('init')}` `{item.get('initName','')}`",
                "",
                "| Trait | Kind | Type / Method |",
                "| --- | --- | --- |",
            ]
        )
        for trait in item.get("traits", []):
            name = trait.get("name", "")
            kind = KIND.get(trait.get("kind"), str(trait.get("kind")))
            detail = ""
            if "type" in trait:
                detail = trait.get("type") or ""
            if "method" in trait:
                method_id = trait["method"]
                method = methods.get(method_id, {})
                detail = f"{method_id} {method.get('name','')}".strip()
            lines.append(f"| `{name}` | `{kind}` | `{detail}` |")
        lines.append("")
        lines.extend(["### Method Hints", ""])
        for trait in item.get("traits", []):
            if "method" not in trait:
                continue
            method_id = trait["method"]
            body = bodies.get(method_id)
            if not body:
                continue
            hints = body.get("bytecodeHints", {})
            if not any(hints.get(key) for key in ("strings", "numbers", "multinames", "methods")):
                continue
            lines.append(f"#### {trait.get('name')} -> method {method_id}")
            if hints.get("numbers"):
                lines.append(f"- Numbers: `{', '.join(map(str, hints['numbers'][:80]))}`")
            if hints.get("strings"):
                lines.append("- Strings: " + ", ".join(f"`{s}`" for s in hints["strings"][:80]))
            if hints.get("multinames"):
                lines.append("- Calls/refs: " + ", ".join(f"`{s}`" for s in hints["multinames"][:100]))
            if hints.get("methods"):
                lines.append("- Method refs: " + ", ".join(f"`{s}`" for s in hints["methods"][:60]))
            lines.append("")

    OUT.write_text("\n".join(lines), encoding="utf-8")
    print(OUT)


if __name__ == "__main__":
    main()
