# Login + Queue + Role Create UI Closed Loop (2026-05-02)

## Goal

Close the first-entry journey in the frontend with legacy semantics:

1. Login announcement
2. Legacy login request
3. Queue wait/polling
4. Role creation for `state=3` users
5. Enter stage when city is ready

## Implemented behavior

- Login page now calls legacy endpoints:
  - `GET /api/legacy/login/announcement`
  - `POST /api/legacy/login`
  - `POST /api/legacy/login/queue` (polling)
  - `POST /api/legacy/role/create`
- Queue state is rendered in-page and polled every 2.5s.
- Stage unlock is strict:
  - only unlock when `sessionUser.cityCount > 0`.
- When logged in but city count is zero, role creation form is shown directly.
- Legacy `uid/sid` auth is persisted in `sessionStorage` to survive refresh during the role-creation step.

## Files

- `frontend/src/layouts/AppLayout.vue`
- `frontend/src/stores/game.ts`
- `frontend/src/api/client.ts`

## Verification

- `npm run build` (frontend) passed.
- `go build -mod=mod ./cmd/api` (backend) passed.
- API smoke chain passed:
  - legacy login returns `raw=[1,2,uid,sid]`
  - role create returns non-zero `cid`.

