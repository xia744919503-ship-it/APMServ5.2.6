# Phase 2 Role Create Compatibility

## Goal

Implement the legacy `UserFunc.php::createRole` flow in Go, with the same visible behavior and the same critical side effects:

- input validation semantics
- starter city allocation + creation
- user state/profile transition
- starter mail/task/gift bootstrap

## Endpoint

- `POST /api/legacy/role/create`

## Request

```json
{
  "uid": 1006,
  "sid": 1569814064,
  "userName": "u23361",
  "cityName": "c45782",
  "province": 0,
  "flagChar": "A",
  "sex": 0,
  "face": 0,
  "code": ""
}
```

## Response

Success example:

```json
{
  "raw": [],
  "uid": 1006,
  "cid": 471185,
  "user": {
    "uid": 1006,
    "name": "u23361"
  }
}
```

Failure example:

```json
{
  "raw": [0, "used_city_holder_name"],
  "uid": 1006,
  "cid": 0
}
```

## Covered parity rules

- session auth:
  - require valid `uid + sid` in `sys_sessions`
- duplicate guards:
  - `sys_user.state != 3` => `cant_duplicate_create`
  - existing city for uid => `cant_duplicate_create`
- validation:
  - username length / city name length
  - illegal chars (legacy-compatible split between username and city name)
  - banned names (`cfg_baned_name`)
  - flag char required and must be single rune
- uniqueness:
  - username already used by other uid => `used_city_holder_name`
- starter city:
  - pick free land from `mem_world` and claim atomically
  - create `sys_city`, level-1 government hall in `sys_building`
  - initialize resource and schedule tables
- profile and bootstrap:
  - update `sys_user` (`state`, `lastcid`, `name`, `face`, `sex`, `flagchar`)
  - seed starter system mails
  - seed first task in `sys_user_task`
  - seed starter gift in `sys_goods` (`gid=50101`)

## Main table side effects

- `mem_world`
- `sys_city`
- `sys_building`
- `sys_city_res_add`
- `mem_city_resource`
- `mem_city_schedule`
- `mem_user_schedule`
- `sys_user`
- `sys_mail_sys_box`
- `sys_alarm`
- `sys_user_task`
- `sys_goods`

## Verification result (2026-05-01)

Smoke chain passed on local runtime:

1. `POST /api/legacy/login` with a new passport returned `raw=[1,2,uid,sid]`.
2. `POST /api/legacy/role/create` with that `uid/sid` returned `raw=[]` and non-zero `cid`.
3. Returned `user.name` matches requested `userName`.

