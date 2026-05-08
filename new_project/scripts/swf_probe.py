from __future__ import annotations

import argparse
import json
import struct
import zlib
from collections import Counter
from pathlib import Path
from typing import Any


TAG_NAMES = {
    0: "End",
    1: "ShowFrame",
    2: "DefineShape",
    4: "PlaceObject",
    5: "RemoveObject",
    6: "DefineBitsJPEG",
    7: "DefineButton",
    8: "JPEGTables",
    9: "SetBackgroundColor",
    10: "DefineFont",
    11: "DefineText",
    12: "DoAction",
    13: "DefineFontInfo",
    14: "DefineSound",
    15: "StartSound",
    17: "DefineButtonSound",
    18: "SoundStreamHead",
    19: "SoundStreamBlock",
    20: "DefineBitsLossless",
    21: "DefineBitsJPEG2",
    22: "DefineShape2",
    23: "DefineButtonCxform",
    24: "Protect",
    26: "PlaceObject2",
    28: "RemoveObject2",
    32: "DefineShape3",
    33: "DefineText2",
    34: "DefineButton2",
    35: "DefineBitsJPEG3",
    36: "DefineBitsLossless2",
    37: "DefineEditText",
    39: "DefineSprite",
    43: "FrameLabel",
    46: "DefineMorphShape",
    48: "DefineFont2",
    56: "ExportAssets",
    57: "ImportAssets",
    58: "EnableDebugger",
    59: "DoInitAction",
    60: "DefineVideoStream",
    61: "VideoFrame",
    64: "EnableDebugger2",
    65: "ScriptLimits",
    66: "SetTabIndex",
    69: "FileAttributes",
    70: "PlaceObject3",
    71: "ImportAssets2",
    73: "DefineFontAlignZones",
    74: "CSMTextSettings",
    75: "DefineFont3",
    76: "SymbolClass",
    77: "Metadata",
    78: "DefineScalingGrid",
    82: "DoABC",
    83: "DefineShape4",
    84: "DefineMorphShape2",
    86: "DefineSceneAndFrameLabelData",
    87: "DefineBinaryData",
    88: "DefineFontName",
    89: "StartSound2",
    90: "DefineBitsJPEG4",
    91: "DefineFont4",
}


ID_TAGS = {
    2,
    6,
    7,
    10,
    11,
    13,
    14,
    20,
    21,
    22,
    32,
    33,
    34,
    35,
    36,
    37,
    39,
    46,
    48,
    60,
    75,
    83,
    84,
    87,
    90,
    91,
}


def read_swf(path: Path) -> tuple[dict[str, Any], bytes]:
    raw = path.read_bytes()
    signature = raw[:3].decode("ascii", errors="replace")
    version = raw[3]
    declared_len = struct.unpack_from("<I", raw, 4)[0]
    if signature == "CWS":
        body = zlib.decompress(raw[8:])
    elif signature == "FWS":
        body = raw[8:]
    else:
        raise ValueError(f"unsupported SWF signature {signature!r}")
    meta = {
        "path": str(path),
        "signature": signature,
        "version": version,
        "declaredLength": declared_len,
        "actualLength": len(body) + 8,
    }
    return meta, body


def rect_size(data: bytes, offset: int = 0) -> int:
    nbits = data[offset] >> 3
    total_bits = 5 + nbits * 4
    return (total_bits + 7) // 8


def cstring(data: bytes, offset: int) -> tuple[str, int]:
    end = data.find(b"\x00", offset)
    if end < 0:
        end = len(data)
    return data[offset:end].decode("utf-8", errors="replace"), end + 1


def parse_tag_stream(body: bytes) -> list[dict[str, Any]]:
    offset = rect_size(body)
    offset += 4  # frame rate + frame count
    tags: list[dict[str, Any]] = []
    while offset + 2 <= len(body):
        start = offset
        header = struct.unpack_from("<H", body, offset)[0]
        offset += 2
        code = header >> 6
        length = header & 0x3F
        if length == 0x3F:
            length = struct.unpack_from("<I", body, offset)[0]
            offset += 4
        payload_start = offset
        payload_end = offset + length
        payload = body[payload_start:payload_end]
        offset = payload_end

        tag: dict[str, Any] = {
            "index": len(tags),
            "code": code,
            "name": TAG_NAMES.get(code, f"Tag{code}"),
            "offset": start,
            "length": length,
        }
        if code in ID_TAGS and len(payload) >= 2:
            tag["id"] = struct.unpack_from("<H", payload, 0)[0]
        if code == 43:
            tag["label"], _ = cstring(payload, 0)
        elif code == 56:
            tag["exports"] = parse_id_name_pairs(payload)
        elif code == 76:
            tag["symbols"] = parse_id_name_pairs(payload)
        elif code == 82:
            flags = struct.unpack_from("<I", payload, 0)[0] if len(payload) >= 4 else 0
            name, abc_offset = cstring(payload, 4)
            abc_data = payload[abc_offset:]
            strings = parse_abc_strings(abc_data)
            abc_summary = parse_abc_summary(abc_data)
            tag["abcName"] = name
            tag["flags"] = flags
            tag["stringCount"] = len(strings)
            tag["interestingStrings"] = interesting_strings(strings)
            tag["abcSummary"] = compact_abc_summary(abc_summary)
        elif code == 87 and len(payload) >= 6:
            tag["binaryReserved"] = struct.unpack_from("<I", payload, 2)[0]
        elif code == 39 and len(payload) >= 4:
            sprite_id, frame_count = struct.unpack_from("<HH", payload, 0)
            tag["id"] = sprite_id
            tag["frameCount"] = frame_count
            tag["placements"] = parse_sprite_placements(payload[4:], sprite_id)
        tags.append(tag)
        if code == 0:
            break
    return tags


def compact_abc_summary(summary: dict[str, Any]) -> dict[str, Any]:
    class_needles = (
        "Login",
        "CreateRole",
        "Jackaroo",
        "BuildingGrid",
        "CreateBuilding",
        "BuildingList",
        "BuildingTip",
        "GuideTip",
        "TopPanel",
        "BottomPanel",
        "InfoPanel",
        "BuildingItem",
        "CityField",
        "CityList",
    )
    interesting_instances = []
    for item in summary.get("instances", []):
        text = f"{item.get('name','')} {item.get('super','')}"
        if any(needle in text for needle in class_needles):
            interesting_instances.append(item)
    interesting_methods = []
    method_needles = (
        "login",
        "create",
        "build",
        "city",
        "guide",
        "init",
        "click",
        "complete",
        "Login",
        "Create",
        "Build",
        "City",
        "Guide",
    )
    for item in summary.get("methods", []):
        name = item.get("name", "")
        if name and any(needle in name for needle in method_needles):
            interesting_methods.append(item)
    return {
        "parseError": summary.get("parseError", ""),
        "instanceCount": len(summary.get("instances", [])),
        "methodCount": len(summary.get("methods", [])),
        "scriptCount": len(summary.get("scripts", [])),
        "interestingInstances": interesting_instances[:400],
        "interestingMethods": interesting_methods[:800],
        "methodBodies": [
            body
            for body in summary.get("methodBodies", [])
            if body.get("bytecodeHints", {}).get("strings")
            or body.get("bytecodeHints", {}).get("numbers")
            or body.get("bytecodeHints", {}).get("multinames")
        ][:3000],
    }


class BitReader:
    def __init__(self, data: bytes, offset: int = 0):
        self.data = data
        self.bit = offset * 8

    def read_bits(self, count: int) -> int:
        value = 0
        for _ in range(count):
            byte = self.data[self.bit // 8]
            shift = 7 - (self.bit % 8)
            value = (value << 1) | ((byte >> shift) & 1)
            self.bit += 1
        return value

    def read_signed(self, count: int) -> int:
        value = self.read_bits(count)
        sign = 1 << (count - 1)
        if value & sign:
            value -= 1 << count
        return value

    def byte_offset(self) -> int:
        return (self.bit + 7) // 8


def parse_matrix(data: bytes, offset: int) -> tuple[dict[str, float], int]:
    r = BitReader(data, offset)
    scale_x = 1.0
    scale_y = 1.0
    rotate_skew0 = 0.0
    rotate_skew1 = 0.0
    if r.read_bits(1):
        bits = r.read_bits(5)
        scale_x = r.read_signed(bits) / 65536
        scale_y = r.read_signed(bits) / 65536
    if r.read_bits(1):
        bits = r.read_bits(5)
        rotate_skew0 = r.read_signed(bits) / 65536
        rotate_skew1 = r.read_signed(bits) / 65536
    translate_bits = r.read_bits(5)
    tx = r.read_signed(translate_bits) / 20 if translate_bits else 0
    ty = r.read_signed(translate_bits) / 20 if translate_bits else 0
    return {
        "x": round(tx, 2),
        "y": round(ty, 2),
        "scaleX": round(scale_x, 4),
        "scaleY": round(scale_y, 4),
        "skew0": round(rotate_skew0, 4),
        "skew1": round(rotate_skew1, 4),
    }, r.byte_offset()


def parse_sprite_placements(data: bytes, sprite_id: int) -> list[dict[str, Any]]:
    placements: list[dict[str, Any]] = []
    offset = 0
    nested_index = 0
    while offset + 2 <= len(data):
        header = struct.unpack_from("<H", data, offset)[0]
        offset += 2
        code = header >> 6
        length = header & 0x3F
        if length == 0x3F:
            if offset + 4 > len(data):
                break
            length = struct.unpack_from("<I", data, offset)[0]
            offset += 4
        payload = data[offset : offset + length]
        offset += length
        if code == 0:
            break
        placement = parse_placement_payload(code, payload)
        if placement:
            placement["spriteId"] = sprite_id
            placement["nestedIndex"] = nested_index
            placements.append(placement)
        nested_index += 1
    return placements


def parse_placement_payload(code: int, payload: bytes) -> dict[str, Any] | None:
    try:
        if code == 4 and len(payload) >= 4:
            character_id, depth = struct.unpack_from("<HH", payload, 0)
            matrix = None
            if len(payload) > 4:
                matrix, _ = parse_matrix(payload, 4)
            return {
                "tag": "PlaceObject",
                "characterId": character_id,
                "depth": depth,
                "matrix": matrix,
            }
        if code in (26, 70) and len(payload) >= (3 if code == 26 else 4):
            flags = payload[0]
            offset = 1
            flags2 = 0
            if code == 70:
                flags2 = payload[1]
                offset = 2
            depth = struct.unpack_from("<H", payload, offset)[0]
            offset += 2
            class_name = ""
            if code == 70 and (flags2 & 0x08 or flags2 & 0x10):
                class_name, offset = cstring(payload, offset)
            character_id = None
            if flags & 0x02:
                character_id = struct.unpack_from("<H", payload, offset)[0]
                offset += 2
            matrix = None
            if flags & 0x04:
                matrix, offset = parse_matrix(payload, offset)
            name = ""
            if flags & 0x20:
                # Matrix, color transform, ratio and clipDepth may appear before name.
                # This lightweight probe records names only when the stream is still aligned.
                try:
                    name, _ = cstring(payload, offset)
                except Exception:
                    name = ""
            return {
                "tag": TAG_NAMES.get(code, f"Tag{code}"),
                "flags": flags,
                "flags2": flags2,
                "characterId": character_id,
                "depth": depth,
                "matrix": matrix,
                "className": class_name,
                "name": name,
            }
    except Exception:
        return None
    return None


def parse_id_name_pairs(payload: bytes) -> list[dict[str, Any]]:
    if len(payload) < 2:
        return []
    count = struct.unpack_from("<H", payload, 0)[0]
    offset = 2
    pairs: list[dict[str, Any]] = []
    for _ in range(count):
        if offset + 2 > len(payload):
            break
        item_id = struct.unpack_from("<H", payload, offset)[0]
        offset += 2
        name, offset = cstring(payload, offset)
        pairs.append({"id": item_id, "name": name})
    return pairs


class Reader:
    def __init__(self, data: bytes):
        self.data = data
        self.i = 0

    def u8(self) -> int:
        value = self.data[self.i]
        self.i += 1
        return value

    def u16(self) -> int:
        value = struct.unpack_from("<H", self.data, self.i)[0]
        self.i += 2
        return value

    def s32(self) -> int:
        value = struct.unpack_from("<i", self.data, self.i)[0]
        self.i += 4
        return value

    def u32(self) -> int:
        value = struct.unpack_from("<I", self.data, self.i)[0]
        self.i += 4
        return value

    def d64(self) -> float:
        value = struct.unpack_from("<d", self.data, self.i)[0]
        self.i += 8
        return value

    def u30(self) -> int:
        result = 0
        for shift in range(0, 35, 7):
            byte = self.u8()
            result |= (byte & 0x7F) << shift
            if not byte & 0x80:
                return result
        return result

    def skip_u30_list(self) -> None:
        count = self.u30()
        for _ in range(1, count):
            self.u30()


def parse_abc_strings(data: bytes) -> list[str]:
    r = Reader(data)
    try:
        r.u16()
        r.u16()
        int_count = r.u30()
        for _ in range(1, int_count):
            r.u30()
        uint_count = r.u30()
        for _ in range(1, uint_count):
            r.u30()
        double_count = r.u30()
        for _ in range(1, double_count):
            r.d64()
        string_count = r.u30()
        strings = [""]
        for _ in range(1, string_count):
            size = r.u30()
            raw = r.data[r.i : r.i + size]
            r.i += size
            strings.append(raw.decode("utf-8", errors="replace"))
        return strings[1:]
    except Exception:
        return []


def parse_abc_summary(data: bytes) -> dict[str, Any]:
    r = Reader(data)
    summary: dict[str, Any] = {
        "strings": [],
        "namespaces": [],
        "multinames": [],
        "instances": [],
        "classes": [],
        "methods": [],
        "scripts": [],
    }
    try:
        r.u16()
        r.u16()
        int_count = r.u30()
        for _ in range(1, int_count):
            r.u30()
        uint_count = r.u30()
        for _ in range(1, uint_count):
            r.u30()
        double_count = r.u30()
        for _ in range(1, double_count):
            r.d64()

        string_count = r.u30()
        strings = [""]
        for _ in range(1, string_count):
            size = r.u30()
            raw = r.data[r.i : r.i + size]
            r.i += size
            strings.append(raw.decode("utf-8", errors="replace"))
        summary["strings"] = strings

        namespace_count = r.u30()
        namespaces: list[dict[str, Any]] = [{"kind": 0, "nameIndex": 0, "name": ""}]
        for _ in range(1, namespace_count):
            kind = r.u8()
            name_index = r.u30()
            namespaces.append(
                {
                    "kind": kind,
                    "nameIndex": name_index,
                    "name": strings[name_index] if name_index < len(strings) else "",
                }
            )
        summary["namespaces"] = namespaces

        ns_set_count = r.u30()
        ns_sets: list[list[int]] = [[]]
        for _ in range(1, ns_set_count):
            count = r.u30()
            ns_sets.append([r.u30() for _ in range(count)])

        multiname_count = r.u30()
        multinames: list[dict[str, Any]] = [{"kind": 0, "name": "", "qname": ""}]
        for _ in range(1, multiname_count):
            multinames.append(parse_multiname(r, strings, namespaces, ns_sets))
        summary["multinames"] = multinames

        method_count = r.u30()
        methods = []
        for index in range(method_count):
            param_count = r.u30()
            return_type = r.u30()
            params = [r.u30() for _ in range(param_count)]
            name_index = r.u30()
            flags = r.u8()
            options = []
            if flags & 0x08:
                option_count = r.u30()
                for _ in range(option_count):
                    options.append({"val": r.u30(), "kind": r.u8()})
            param_names = []
            if flags & 0x80:
                param_names = [r.u30() for _ in range(param_count)]
            methods.append(
                {
                    "index": index,
                    "name": strings[name_index] if name_index < len(strings) else "",
                    "return": multiname_name(multinames, return_type),
                    "params": [multiname_name(multinames, item) for item in params],
                    "paramNames": [strings[item] if item < len(strings) else "" for item in param_names],
                    "flags": flags,
                }
            )
        summary["methods"] = methods

        metadata_count = r.u30()
        for _ in range(metadata_count):
            r.u30()
            item_count = r.u30()
            for _ in range(item_count):
                r.u30()
                r.u30()

        instance_count = r.u30()
        instances = []
        for index in range(instance_count):
            name = r.u30()
            super_name = r.u30()
            flags = r.u8()
            protected_ns = None
            if flags & 0x08:
                protected_ns = r.u30()
            interface_count = r.u30()
            interfaces = [r.u30() for _ in range(interface_count)]
            init = r.u30()
            traits = parse_traits(r, strings, multinames)
            instances.append(
                {
                    "index": index,
                    "name": multiname_name(multinames, name),
                    "super": multiname_name(multinames, super_name),
                    "flags": flags,
                    "protectedNs": namespace_name(namespaces, protected_ns),
                    "interfaces": [multiname_name(multinames, item) for item in interfaces],
                    "init": init,
                    "initName": methods[init]["name"] if init < len(methods) else "",
                    "traits": traits,
                }
            )
        summary["instances"] = instances

        classes = []
        for index in range(instance_count):
            cinit = r.u30()
            traits = parse_traits(r, strings, multinames)
            classes.append(
                {
                    "index": index,
                    "name": instances[index]["name"] if index < len(instances) else "",
                    "cinit": cinit,
                    "traits": traits,
                }
            )
        summary["classes"] = classes

        script_count = r.u30()
        scripts = []
        for index in range(script_count):
            init = r.u30()
            traits = parse_traits(r, strings, multinames)
            scripts.append(
                {
                    "index": index,
                    "init": init,
                    "traits": traits,
                }
            )
        summary["scripts"] = scripts

        method_body_count = r.u30()
        bodies = []
        for _ in range(method_body_count):
            method = r.u30()
            max_stack = r.u30()
            local_count = r.u30()
            init_scope_depth = r.u30()
            max_scope_depth = r.u30()
            code_length = r.u30()
            code_start = r.i
            code = r.data[code_start : code_start + code_length]
            r.i += code_length
            exception_count = r.u30()
            for _ in range(exception_count):
                r.u30()
                r.u30()
                r.u30()
                r.u30()
                r.u30()
            traits = parse_traits(r, strings, multinames)
            bodies.append(
                {
                    "method": method,
                    "methodName": methods[method]["name"] if method < len(methods) else "",
                    "codeLength": code_length,
                    "maxStack": max_stack,
                    "localCount": local_count,
                    "traits": traits,
                    "codeStart": code_start,
                    "bytecodeHints": decode_bytecode_hints(code, summary, methods),
                    "initScopeDepth": init_scope_depth,
                    "maxScopeDepth": max_scope_depth,
                }
            )
        summary["methodBodies"] = bodies
    except Exception as exc:
        summary["parseError"] = str(exc)
    return summary


def decode_bytecode_hints(code: bytes, summary: dict[str, Any], methods: list[dict[str, Any]]) -> dict[str, Any]:
    strings = summary.get("strings", [])
    multinames = summary.get("multinames", [])
    ints: list[int] = []
    uints: list[int] = []
    doubles: list[int] = []
    string_refs: list[str] = []
    multiname_refs: list[str] = []
    method_refs: list[str] = []
    small_numbers: list[int] = []
    i = 0
    while i < len(code):
        op = code[i]
        i += 1
        try:
            if op == 0x24:  # pushbyte
                value = struct.unpack_from("b", code, i)[0]
                i += 1
                small_numbers.append(value)
            elif op == 0x25:  # pushshort
                value, i = read_u30_from_bytes(code, i)
                small_numbers.append(value)
            elif op == 0x2C:  # pushstring
                index, i = read_u30_from_bytes(code, i)
                if index < len(strings):
                    string_refs.append(strings[index])
            elif op == 0x2D:  # pushint
                index, i = read_u30_from_bytes(code, i)
                ints.append(index)
            elif op == 0x2E:  # pushuint
                index, i = read_u30_from_bytes(code, i)
                uints.append(index)
            elif op == 0x2F:  # pushdouble
                index, i = read_u30_from_bytes(code, i)
                doubles.append(index)
            elif op in (0x40,):  # newfunction
                index, i = read_u30_from_bytes(code, i)
                if index < len(methods):
                    method_refs.append(methods[index].get("name", ""))
            elif op in (0x46, 0x4A):  # callproperty / constructprop
                index, i = read_u30_from_bytes(code, i)
                argc, i = read_u30_from_bytes(code, i)
                if index < len(multinames):
                    multiname_refs.append(multinames[index].get("qname", ""))
            elif op in (0x47, 0x48):  # returnvoid/value
                pass
            elif op in (0x49,):  # constructsuper
                _, i = read_u30_from_bytes(code, i)
            elif op in (0x4F, 0x80):  # callpropvoid / coerce
                index, i = read_u30_from_bytes(code, i)
                if op == 0x4F:
                    _, i = read_u30_from_bytes(code, i)
                if index < len(multinames):
                    multiname_refs.append(multinames[index].get("qname", ""))
            elif op in (0x55,):  # newobject
                _, i = read_u30_from_bytes(code, i)
            elif op in (0x56,):  # newarray
                _, i = read_u30_from_bytes(code, i)
            elif op in (0x58,):  # newclass
                _, i = read_u30_from_bytes(code, i)
            elif op in (0x5D, 0x5E, 0x60, 0x61, 0x66, 0x68):  # find/get/set property lexical
                index, i = read_u30_from_bytes(code, i)
                if index < len(multinames):
                    multiname_refs.append(multinames[index].get("qname", ""))
            elif op in (0x62, 0x63):  # getlocal/setlocal
                _, i = read_u30_from_bytes(code, i)
            elif op in (0x64, 0x65):  # getglobalscope/getscopeobject
                if op == 0x65:
                    i += 1
            elif op in (0x6A,):  # deleteproperty
                index, i = read_u30_from_bytes(code, i)
                if index < len(multinames):
                    multiname_refs.append(multinames[index].get("qname", ""))
            elif op in (0xD0, 0xD1, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6, 0xD7):
                pass
            elif op in (0xD8, 0xD9, 0xDA, 0xDB):
                pass
            elif op in (0xEF,):  # debug
                i += 1
                _, i = read_u30_from_bytes(code, i)
                i += 1
                _, i = read_u30_from_bytes(code, i)
            elif op in (0xF0,):  # debugline
                _, i = read_u30_from_bytes(code, i)
            elif op in (0xF1,):  # debugfile
                _, i = read_u30_from_bytes(code, i)
            elif op in (0x10, 0x11, 0x12, 0x13, 0x14):  # jumps
                i += 3
            elif op in (0x1A,):  # lookupswitch
                i += 3
                case_count, i = read_u30_from_bytes(code, i)
                i += 3 * (case_count + 1)
            elif op in (0x08,):  # kill
                _, i = read_u30_from_bytes(code, i)
            elif op in (0x09,):  # label
                pass
            elif op in (0x1C, 0x1D, 0x1E, 0x1F):  # pushwith/popscope/next...
                pass
            elif op in (0xC0, 0xC1, 0xC2, 0xC3):  # inclocal_i etc simplified
                pass
            else:
                # Most arithmetic and stack opcodes are single byte. Unknown operands will desync
                # only for uncommon opcodes; hints are best-effort.
                pass
        except Exception:
            break
    return {
        "numbers": sorted(set(n for n in small_numbers if -2000 <= n <= 5000))[:200],
        "strings": sorted(set(s for s in string_refs if s))[:200],
        "multinames": sorted(set(s for s in multiname_refs if s))[:240],
        "methods": sorted(set(s for s in method_refs if s))[:80],
        "intIndexes": sorted(set(ints))[:80],
        "uintIndexes": sorted(set(uints))[:80],
        "doubleIndexes": sorted(set(doubles))[:80],
    }


def read_u30_from_bytes(data: bytes, offset: int) -> tuple[int, int]:
    result = 0
    shift = 0
    while True:
        byte = data[offset]
        offset += 1
        result |= (byte & 0x7F) << shift
        if not byte & 0x80:
            return result, offset
        shift += 7


def namespace_name(namespaces: list[dict[str, Any]], index: int | None) -> str:
    if index is None or index >= len(namespaces):
        return ""
    return namespaces[index].get("name", "")


def multiname_name(multinames: list[dict[str, Any]], index: int) -> str:
    if 0 <= index < len(multinames):
        return multinames[index].get("qname") or multinames[index].get("name") or ""
    return ""


def parse_multiname(
    r: Reader,
    strings: list[str],
    namespaces: list[dict[str, Any]],
    ns_sets: list[list[int]],
) -> dict[str, Any]:
    kind = r.u8()
    item: dict[str, Any] = {"kind": kind, "name": "", "qname": ""}
    if kind in (0x07, 0x0D):
        ns_index = r.u30()
        name_index = r.u30()
        ns = namespace_name(namespaces, ns_index)
        name = strings[name_index] if name_index < len(strings) else ""
        item.update({"namespace": ns, "name": name, "qname": f"{ns}.{name}" if ns else name})
    elif kind in (0x0F, 0x10):
        name_index = r.u30()
        name = strings[name_index] if name_index < len(strings) else ""
        item.update({"name": name, "qname": name})
    elif kind in (0x11, 0x12):
        item.update({"name": "*", "qname": "*"})
    elif kind in (0x09, 0x0E):
        name_index = r.u30()
        ns_set_index = r.u30()
        name = strings[name_index] if name_index < len(strings) else ""
        ns_names = []
        if ns_set_index < len(ns_sets):
            ns_names = [namespace_name(namespaces, ns) for ns in ns_sets[ns_set_index]]
        item.update({"name": name, "namespaces": ns_names, "qname": f"{'|'.join(ns_names)}::{name}"})
    elif kind in (0x1B, 0x1C):
        ns_set_index = r.u30()
        ns_names = []
        if ns_set_index < len(ns_sets):
            ns_names = [namespace_name(namespaces, ns) for ns in ns_sets[ns_set_index]]
        item.update({"name": "*", "namespaces": ns_names, "qname": f"{'|'.join(ns_names)}::*"})
    elif kind == 0x1D:
        qname_index = r.u30()
        param_count = r.u30()
        params = [r.u30() for _ in range(param_count)]
        item.update(
            {
                "name": multiname_name([{}] + [], qname_index),
                "qnameIndex": qname_index,
                "params": params,
            }
        )
    return item


def parse_traits(r: Reader, strings: list[str], multinames: list[dict[str, Any]]) -> list[dict[str, Any]]:
    traits = []
    trait_count = r.u30()
    for _ in range(trait_count):
        name_index = r.u30()
        kind_attr = r.u8()
        kind = kind_attr & 0x0F
        attrs = kind_attr >> 4
        trait: dict[str, Any] = {
            "name": multiname_name(multinames, name_index),
            "kind": kind,
            "attrs": attrs,
        }
        if kind in (0, 6):
            trait["slotId"] = r.u30()
            trait["type"] = multiname_name(multinames, r.u30())
            vindex = r.u30()
            trait["vindex"] = vindex
            if vindex:
                trait["vkind"] = r.u8()
        elif kind in (1, 2, 3):
            trait["dispId"] = r.u30()
            trait["method"] = r.u30()
        elif kind == 4:
            trait["slotId"] = r.u30()
            trait["classi"] = r.u30()
        elif kind == 5:
            trait["slotId"] = r.u30()
            trait["function"] = r.u30()
        if attrs & 0x04:
            metadata_count = r.u30()
            trait["metadata"] = [r.u30() for _ in range(metadata_count)]
        traits.append(trait)
    return traits


def interesting_strings(strings: list[str]) -> list[str]:
    needles = (
        "login",
        "role",
        "city",
        "inner",
        "build",
        "guide",
        "task",
        "map",
        "war",
        "hero",
        "Login",
        "Role",
        "City",
        "Build",
        "Guide",
        "Main",
        "Panel",
        "NetConnection",
        "RemoteObject",
        "amf",
        "create",
        "province",
        "bloodwar",
        "热",
        "城",
        "建",
        "州",
        "新手",
    )
    found = []
    for item in strings:
        if 2 <= len(item) <= 140 and any(needle in item for needle in needles):
            found.append(item)
    return sorted(set(found))[:300]


def summarize(path: Path) -> dict[str, Any]:
    meta, body = read_swf(path)
    tags = parse_tag_stream(body)
    tag_counts = Counter(tag["name"] for tag in tags)
    symbols = [pair for tag in tags for pair in tag.get("symbols", [])]
    exports = [pair for tag in tags for pair in tag.get("exports", [])]
    abc = [tag for tag in tags if tag["code"] == 82]
    interesting = sorted(set(s for tag in abc for s in tag.get("interestingStrings", [])))
    id_tags = [
        {"id": tag["id"], "name": tag["name"], "index": tag["index"], "length": tag["length"]}
        for tag in tags
        if "id" in tag
    ]
    symbol_by_id = {pair["id"]: pair["name"] for pair in symbols}
    placements = []
    for tag in tags:
        for placement in tag.get("placements", []):
            item = dict(placement)
            character_id = item.get("characterId")
            if character_id is not None:
                item["characterName"] = symbol_by_id.get(character_id, "")
            item["parentName"] = symbol_by_id.get(item.get("spriteId"), "")
            placements.append(item)
    return {
        "meta": meta,
        "tagCounts": dict(tag_counts.most_common()),
        "symbols": symbols,
        "exports": exports,
        "abcBlocks": [
            {
                "index": tag["index"],
                "abcName": tag.get("abcName", ""),
                "stringCount": tag.get("stringCount", 0),
                "interestingStrings": tag.get("interestingStrings", []),
                "abcSummary": tag.get("abcSummary", {}),
            }
            for tag in abc
        ],
        "interestingStrings": interesting,
        "idTags": id_tags,
        "placements": placements,
        "tags": tags,
    }


def write_markdown(result: dict[str, Any], out: Path) -> None:
    meta = result["meta"]
    lines = [
        f"# SWF Probe: {Path(meta['path']).name}",
        "",
        f"- Path: `{meta['path']}`",
        f"- Signature: `{meta['signature']}`",
        f"- Version: `{meta['version']}`",
        f"- Declared length: `{meta['declaredLength']}`",
        "",
        "## Tag Counts",
        "",
    ]
    for name, count in result["tagCounts"].items():
        lines.append(f"- `{name}`: {count}")
    lines.extend(["", "## SymbolClass", ""])
    for pair in result["symbols"][:500]:
        lines.append(f"- `{pair['id']}` -> `{pair['name']}`")
    if len(result["symbols"]) > 500:
        lines.append(f"- ... {len(result['symbols']) - 500} more")
    lines.extend(["", "## ExportAssets", ""])
    for pair in result["exports"][:500]:
        lines.append(f"- `{pair['id']}` -> `{pair['name']}`")
    lines.extend(["", "## Interesting ABC Strings", ""])
    for item in result["interestingStrings"][:1000]:
        lines.append(f"- `{item}`")
    lines.extend(["", "## Interesting ABC Classes", ""])
    for block in result.get("abcBlocks", []):
        summary = block.get("abcSummary", {})
        if summary.get("parseError"):
            lines.append(f"- ABC `{block.get('abcName')}` parse error: `{summary.get('parseError')}`")
        for item in summary.get("interestingInstances", []):
            trait_names = ", ".join(t.get("name", "") for t in item.get("traits", [])[:24] if t.get("name"))
            lines.append(
                f"- `{item.get('name')}` extends `{item.get('super')}` init `{item.get('initName')}` traits `{trait_names}`"
            )
    lines.extend(["", "## Interesting ABC Methods", ""])
    for block in result.get("abcBlocks", []):
        summary = block.get("abcSummary", {})
        for item in summary.get("interestingMethods", [])[:300]:
            params = ", ".join(item.get("params", []))
            lines.append(f"- `{item.get('index')}` `{item.get('name')}`({params}) -> `{item.get('return')}`")
    lines.extend(["", "## Relevant Placements", ""])
    for item in result.get("placements", []):
        haystack = f"{item.get('parentName','')} {item.get('characterName','')} {item.get('name','')}"
        if any(key in haystack for key in ("Login", "CreateRole", "City", "Building", "Guide", "Bar", "Panel", "board_login", "mycity", "leftboard")):
            lines.append(
                f"- parent `{item.get('parentName','') or item.get('spriteId')}` "
                f"places `{item.get('characterName','') or item.get('characterId')}` "
                f"depth `{item.get('depth')}` matrix `{item.get('matrix')}`"
            )
    out.write_text("\n".join(lines) + "\n", encoding="utf-8")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("swf", nargs="+", type=Path)
    parser.add_argument("--out-dir", type=Path, default=Path("swf-probe-out"))
    args = parser.parse_args()

    args.out_dir.mkdir(parents=True, exist_ok=True)
    for swf in args.swf:
        result = summarize(swf)
        stem = swf.stem
        (args.out_dir / f"{stem}.json").write_text(
            json.dumps(result, ensure_ascii=False, indent=2), encoding="utf-8"
        )
        write_markdown(result, args.out_dir / f"{stem}.md")
        print(f"wrote {args.out_dir / (stem + '.json')}")


if __name__ == "__main__":
    main()
