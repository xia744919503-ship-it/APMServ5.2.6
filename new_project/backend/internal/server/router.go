package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"rxsg-new-project/backend/internal/config"
	"rxsg-new-project/backend/internal/httpjson"
	"rxsg-new-project/backend/internal/legacy"
	"rxsg-new-project/backend/internal/service"
)

type Server struct {
	cfg      config.Config
	service  service.Service
	sessions *sessionStore
}

func New(cfg config.Config, svc service.Service) Server {
	return Server{
		cfg:      cfg,
		service:  svc,
		sessions: newSessionStore(),
	}
}

func (s Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/dashboard/overview", s.handleOverview)
	mux.HandleFunc("/api/legacy/portal", s.handleLegacyPortal)
	mux.HandleFunc("/api/legacy/login/announcement", s.handleLegacyLoginAnnouncement)
	mux.HandleFunc("/api/legacy/login", s.handleLegacyLogin)
	mux.HandleFunc("/api/legacy/login/queue", s.handleLegacyLoginQueue)
	mux.HandleFunc("/api/legacy/role/create", s.handleLegacyRoleCreate)
	mux.HandleFunc("/api/legacy/guides", s.handleLegacyGuides)
	mux.HandleFunc("/api/legacy/activities", s.handleLegacyActivities)
	mux.HandleFunc("/api/legacy/user-type-goods/use", s.handleLegacyUseUserTypeGoods)
	mux.HandleFunc("/api/legacy/user-type-goods", s.handleLegacyUserTypeGoods)
	mux.HandleFunc("/api/auth/commanders", s.handleCommanderOptions)
	mux.HandleFunc("/api/auth/me", s.handleCurrentUser)
	mux.HandleFunc("/api/auth/login", s.handleLogin)
	mux.HandleFunc("/api/auth/logout", s.handleLogout)
	mux.HandleFunc("/api/me/cities", s.handleMyCities)
	mux.HandleFunc("/api/me/relations", s.handleMyRelations)
	mux.HandleFunc("/api/me/relations/remove", s.handleMyRelationRemove)
	mux.HandleFunc("/api/me/union", s.handleMyUnion)
	mux.HandleFunc("/api/me/union/create", s.handleMyUnionCreate)
	mux.HandleFunc("/api/me/union/apply", s.handleMyUnionApply)
	mux.HandleFunc("/api/me/union/apply/cancel", s.handleMyUnionApplyCancel)
	mux.HandleFunc("/api/me/union/leave", s.handleMyUnionLeave)
	mux.HandleFunc("/api/me/union/profile", s.handleMyUnionProfile)
	mux.HandleFunc("/api/me/union/relations", s.handleMyUnionRelationSet)
	mux.HandleFunc("/api/me/union/relations/remove", s.handleMyUnionRelationRemove)
	mux.HandleFunc("/api/me/tasks", s.handleMyTasks)
	mux.HandleFunc("/api/me/tasks/claim", s.handleMyTaskClaim)
	mux.HandleFunc("/api/me/shop", s.handleMyShop)
	mux.HandleFunc("/api/me/shop/buy", s.handleMyShopBuy)
	mux.HandleFunc("/api/me/charge", s.handleMyCharge)
	mux.HandleFunc("/api/me/charge/exchange", s.handleMyChargeExchange)
	mux.HandleFunc("/api/me/troops", s.handleMyTroops)
	mux.HandleFunc("/api/battle/field-state", s.handleBattleFieldState)
	mux.HandleFunc("/api/battle/quit-preview", s.handleBattleQuitPreview)
	mux.HandleFunc("/api/battle/troop-detail", s.handleBattleTroopDetail)
	mux.HandleFunc("/api/battle/army-send-preview", s.handleBattleArmySendPreview)
	mux.HandleFunc("/api/battle/campaign-preview", s.handleBattleCampaignPreview)
	mux.HandleFunc("/api/battle/army-attack-preview", s.handleBattleArmyAttackPreview)
	mux.HandleFunc("/api/battle/patrol-preview", s.handleBattlePatrolPreview)
	mux.HandleFunc("/api/battle/tasks", s.handleBattleTasks)
	mux.HandleFunc("/api/battle/members", s.handleBattleMembers)
	mux.HandleFunc("/api/battle/news", s.handleBattleNews)
	mux.HandleFunc("/api/troops/", s.handleTroopRoute)
	mux.HandleFunc("/api/mail/delete", s.handleMailDelete)
	mux.HandleFunc("/api/mail/send", s.handleMailSend)
	mux.HandleFunc("/api/mail", s.handleMail)
	mux.HandleFunc("/api/mail/", s.handleMailRoute)
	mux.HandleFunc("/api/reports", s.handleReports)
	mux.HandleFunc("/api/reports/", s.handleReportRoute)
	mux.HandleFunc("/api/rankings", s.handleRankings)
	mux.HandleFunc("/api/cities", s.handleCities)
	mux.HandleFunc("/api/cities/", s.handleCityRoute)
	mux.HandleFunc("/api/world/map", s.handleWorldMap)
	mux.HandleFunc("/", s.handleFrontend)
	return withAccessLog(withCORS(mux))
}

func (s Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	httpjson.Write(w, http.StatusOK, s.service.Health())
}

func (s Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	payload, err := s.service.Overview(r.Context())
	if err != nil {
		httpjson.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleLegacyPortal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	payload, err := s.service.LegacyPortal(r.Context())
	if err != nil {
		httpjson.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleLegacyLoginAnnouncement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	announcement, err := s.service.LegacyLoginAnnouncement(r.Context())
	if err != nil {
		httpjson.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpjson.Write(w, http.StatusOK, map[string]any{
		"announcement": announcement,
	})
}

func (s Server) handleLegacyLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var payload legacy.LegacyLoginPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := s.service.LegacyDoLogin(r.Context(), payload, requestIPInt(r))
	if err != nil {
		httpjson.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	if result.Logged && result.UID > 0 {
		if token, err := s.sessions.create(result.UID); err == nil {
			s.writeSessionCookie(w, token)
		}
		if user, err := s.service.SessionUser(r.Context(), result.UID); err == nil {
			result.User = &user
		}
	}

	httpjson.Write(w, http.StatusOK, result)
}

func (s Server) handleLegacyLoginQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var payload legacy.LegacyQueueCheckPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := s.service.LegacyCheckQueue(r.Context(), payload, requestIPInt(r))
	if err != nil {
		httpjson.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	if result.Logged && result.UID > 0 {
		if token, err := s.sessions.create(result.UID); err == nil {
			s.writeSessionCookie(w, token)
		}
		if user, err := s.service.SessionUser(r.Context(), result.UID); err == nil {
			result.User = &user
		}
	}

	httpjson.Write(w, http.StatusOK, result)
}

func (s Server) handleLegacyRoleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var payload legacy.LegacyRoleCreatePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := s.service.LegacyCreateRole(r.Context(), payload)
	if err != nil {
		httpjson.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	if result.UID > 0 {
		if token, tokenErr := s.sessions.create(result.UID); tokenErr == nil {
			s.writeSessionCookie(w, token)
		}
		if user, userErr := s.service.SessionUser(r.Context(), result.UID); userErr == nil {
			result.User = &user
		}
	}

	httpjson.Write(w, http.StatusOK, result)
}

func (s Server) handleLegacyGuides(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	group, err := strconv.Atoi(defaultString(r.URL.Query().Get("group"), "1"))
	if err != nil || group <= 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid guide group")
		return
	}

	payload, err := s.service.GuidesByGroup(r.Context(), group)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleLegacyActivities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	payload, err := s.service.ActivityList(r.Context())
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleLegacyUserTypeGoods(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	goodsType, err := strconv.Atoi(defaultString(r.URL.Query().Get("type"), "-1"))
	if err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid goods type")
		return
	}

	timeLeft, err := strconv.ParseInt(defaultString(r.URL.Query().Get("timeLeft"), "0"), 10, 64)
	if err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid timeLeft")
		return
	}

	payload, err := s.service.UserTypeGoods(r.Context(), uid, goodsType, timeLeft)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleLegacyUseUserTypeGoods(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		Type int `json:"type"`
		GID  int `json:"gid"`
		CID  int `json:"cid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if payload.CID <= 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid city id")
		return
	}
	if payload.GID <= 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid goods id")
		return
	}

	detail, err := s.service.UseUserTypeGoods(r.Context(), uid, payload.CID, payload.Type, payload.GID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, detail)
}

func (s Server) handleCommanderOptions(w http.ResponseWriter, r *http.Request) {
	limit := 18
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	items, err := s.service.CommanderOptions(r.Context(), limit)
	if err != nil {
		httpjson.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpjson.Write(w, http.StatusOK, map[string]any{
		"items": items,
	})
}

func (s Server) handleCurrentUser(w http.ResponseWriter, r *http.Request) {
	uid, ok := s.currentUID(r)
	if !ok {
		httpjson.Write(w, http.StatusOK, map[string]any{
			"user": nil,
		})
		return
	}

	user, err := s.service.SessionUser(r.Context(), uid)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusUnauthorized
		}
		httpjson.Error(w, status, err.Error())
		return
	}

	httpjson.Write(w, http.StatusOK, map[string]any{
		"user": user,
	})
}

func (s Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var payload struct {
		UID      int    `json:"uid"`
		Passport string `json:"passport"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var (
		user legacy.SessionUser
		err  error
	)

	switch {
	case payload.UID > 0:
		user, err = s.service.SessionUser(r.Context(), payload.UID)
	case strings.TrimSpace(payload.Passport) != "":
		user, err = s.service.LoginByPassport(r.Context(), payload.Passport, payload.Password)
	default:
		httpjson.Error(w, http.StatusBadRequest, "uid or passport is required")
		return
	}

	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, sql.ErrNoRows), errors.Is(err, legacy.ErrInvalid):
			status = http.StatusUnauthorized
		case errors.Is(err, legacy.ErrForbidden):
			status = http.StatusForbidden
		}
		httpjson.Error(w, status, err.Error())
		return
	}

	token, err := s.sessions.create(user.UID)
	if err != nil {
		httpjson.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	sid := time.Now().UnixNano() & 0x7fffffff
	_ = s.service.TouchLegacySession(r.Context(), user.UID, sid, requestIPInt(r))
	s.writeSessionCookie(w, token)

	httpjson.Write(w, http.StatusOK, map[string]any{
		"ok":   true,
		"user": user,
	})
}

func (s Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		s.sessions.delete(cookie.Value)
	}
	s.clearSessionCookie(w)

	httpjson.Write(w, http.StatusOK, map[string]any{
		"ok": true,
	})
}

func (s Server) handleMyCities(w http.ResponseWriter, r *http.Request) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	limit := 80
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	items, err := s.service.UserCities(r.Context(), uid, limit)
	if err != nil {
		httpjson.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpjson.Write(w, http.StatusOK, map[string]any{
		"items": items,
	})
}

func (s Server) handleMyRelations(w http.ResponseWriter, r *http.Request) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	switch r.Method {
	case http.MethodGet:
		payload, err := s.service.UserRelations(r.Context(), uid)
		if err != nil {
			s.writeDomainError(w, err)
			return
		}

		httpjson.Write(w, http.StatusOK, payload)
	case http.MethodPost:
		var payload struct {
			Name         string `json:"name"`
			RelationType int    `json:"relationType"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid request body")
			return
		}

		page, err := s.service.AddUserRelation(r.Context(), uid, payload.Name, payload.RelationType)
		if err != nil {
			s.writeDomainError(w, err)
			return
		}

		httpjson.Write(w, http.StatusOK, page)
	default:
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s Server) handleMyRelationRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		TargetUID    int `json:"targetUid"`
		RelationType int `json:"relationType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	page, err := s.service.RemoveUserRelation(r.Context(), uid, payload.TargetUID, payload.RelationType)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, page)
}

func (s Server) handleMyUnion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	payload, err := s.service.MyUnion(r.Context(), uid)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleMyUnionCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.CreateUnion(r.Context(), uid, payload.Name)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleMyUnionApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		UnionID int `json:"unionId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.ApplyJoinUnion(r.Context(), uid, payload.UnionID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleMyUnionApplyCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	snapshot, err := s.service.CancelJoinUnionApply(r.Context(), uid)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleMyUnionLeave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	snapshot, err := s.service.LeaveUnion(r.Context(), uid)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleMyUnionProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		Name         string `json:"name"`
		Intro        string `json:"intro"`
		Announcement string `json:"announcement"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.UpdateUnionProfile(r.Context(), uid, payload.Name, payload.Intro, payload.Announcement)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleMyUnionRelationSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		TargetUnionID int `json:"targetUnionId"`
		RelationType  int `json:"relationType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.SetUnionRelation(r.Context(), uid, payload.TargetUnionID, payload.RelationType)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleMyUnionRelationRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		TargetUnionID int `json:"targetUnionId"`
		RelationType  int `json:"relationType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.RemoveUnionRelation(r.Context(), uid, payload.TargetUnionID, payload.RelationType)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleMyTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	payload, err := s.service.MyTasks(r.Context(), uid)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleMyTaskClaim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		TaskID int `json:"taskId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.ClaimTaskReward(r.Context(), uid, payload.TaskID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleBattleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	bid, err := strconv.Atoi(defaultString(r.URL.Query().Get("bid"), "0"))
	if err != nil || bid < 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid bid")
		return
	}
	unionID, err := strconv.Atoi(defaultString(r.URL.Query().Get("unionId"), "0"))
	if err != nil || unionID < 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid unionId")
		return
	}

	payload, err := s.service.BattleTasksSnapshot(r.Context(), uid, bid, unionID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleMyShop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	payload, err := s.service.MyShop(r.Context(), uid)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleMyShopBuy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		ItemID  int `json:"itemId"`
		Count   int `json:"count"`
		PayType int `json:"payType"`
		CityID  int `json:"cityId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.BuyShopItem(r.Context(), uid, payload.ItemID, payload.Count, payload.PayType, payload.CityID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleMyCharge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	payload, err := s.service.MyCharge(r.Context(), uid)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleMyChargeExchange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		ExchangeCount int `json:"exchangeCount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.ExchangeCharge(r.Context(), uid, payload.ExchangeCount)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleMyTroops(w http.ResponseWriter, r *http.Request) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	limit := 40
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	payload, err := s.service.MyTroops(r.Context(), uid, limit)
	if err != nil {
		httpjson.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleBattleFieldState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	battlefieldID, err := strconv.Atoi(defaultString(r.URL.Query().Get("battlefieldId"), "0"))
	if err != nil || battlefieldID < 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid battlefieldId")
		return
	}
	unionID, err := strconv.Atoi(defaultString(r.URL.Query().Get("unionId"), "0"))
	if err != nil || unionID < 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid unionId")
		return
	}
	cid, err := strconv.Atoi(defaultString(r.URL.Query().Get("cid"), "0"))
	if err != nil || cid < 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid cid")
		return
	}

	payload, err := s.service.BattleFieldState(r.Context(), uid, battlefieldID, unionID, cid, r.URL.Query().Get("name"))
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleBattleQuitPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	payload, err := s.service.BattleQuitPreview(r.Context(), uid)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleBattleTroopDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	troopID, err := strconv.Atoi(defaultString(r.URL.Query().Get("troopId"), "0"))
	if err != nil || troopID <= 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid troopId")
		return
	}

	payload, err := s.service.BattleTroopDetail(r.Context(), uid, troopID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleBattleArmySendPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	troopID, err := strconv.Atoi(defaultString(r.URL.Query().Get("troopId"), "0"))
	if err != nil || troopID <= 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid troopId")
		return
	}
	targetCID, err := strconv.Atoi(defaultString(r.URL.Query().Get("targetCid"), "0"))
	if err != nil || targetCID <= 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid targetCid")
		return
	}

	payload, err := s.service.BattleArmySendPreview(r.Context(), uid, troopID, targetCID, r.URL.Query().Get("targetName"))
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleBattleCampaignPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	cid, err := strconv.Atoi(defaultString(r.URL.Query().Get("cid"), "0"))
	if err != nil || cid <= 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid cid")
		return
	}
	targetCID, err := strconv.Atoi(defaultString(r.URL.Query().Get("targetCid"), "0"))
	if err != nil || targetCID < 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid targetCid")
		return
	}
	heroID, err := strconv.Atoi(defaultString(r.URL.Query().Get("heroId"), "0"))
	if err != nil || heroID < 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid heroId")
		return
	}

	soldiers := map[int]int64{}
	for _, raw := range r.URL.Query()["soldiers"] {
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) != 2 {
			httpjson.Error(w, http.StatusBadRequest, "invalid soldiers")
			return
		}
		sid, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil || sid <= 0 {
			httpjson.Error(w, http.StatusBadRequest, "invalid soldier sid")
			return
		}
		count, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil || count < 0 {
			httpjson.Error(w, http.StatusBadRequest, "invalid soldier count")
			return
		}
		if count > 0 {
			soldiers[sid] += count
		}
	}
	useFlag := strings.EqualFold(r.URL.Query().Get("useFlag"), "true") || r.URL.Query().Get("useFlag") == "1"

	payload, err := s.service.BattleCampaignPreview(r.Context(), uid, cid, targetCID, heroID, soldiers, useFlag)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleBattleArmyAttackPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	troopID, err := strconv.Atoi(defaultString(r.URL.Query().Get("troopId"), "0"))
	if err != nil || troopID <= 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid troopId")
		return
	}
	targetTroopID, err := strconv.Atoi(defaultString(r.URL.Query().Get("targetTroopId"), "0"))
	if err != nil || targetTroopID <= 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid targetTroopId")
		return
	}

	payload, err := s.service.BattleArmyAttackPreview(r.Context(), uid, troopID, targetTroopID, r.URL.Query().Get("targetName"))
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleBattlePatrolPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	troopID, err := strconv.Atoi(defaultString(r.URL.Query().Get("troopId"), "0"))
	if err != nil || troopID <= 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid troopId")
		return
	}
	targetTroopID, err := strconv.Atoi(defaultString(r.URL.Query().Get("targetTroopId"), "0"))
	if err != nil || targetTroopID <= 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid targetTroopId")
		return
	}

	payload, err := s.service.BattlePatrolPreview(r.Context(), uid, troopID, targetTroopID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleBattleMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	payload, err := s.service.BattleMembersSnapshot(r.Context(), uid)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleBattleNews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	battlefieldID, err := strconv.Atoi(defaultString(r.URL.Query().Get("battlefieldId"), "0"))
	if err != nil || battlefieldID < 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid battlefieldId")
		return
	}
	unionID, err := strconv.Atoi(defaultString(r.URL.Query().Get("unionId"), "0"))
	if err != nil || unionID < 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid unionId")
		return
	}
	page, err := strconv.Atoi(defaultString(r.URL.Query().Get("page"), "1"))
	if err != nil || page <= 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid page")
		return
	}
	pageSize, err := strconv.Atoi(defaultString(r.URL.Query().Get("pageSize"), "10"))
	if err != nil || pageSize <= 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid pageSize")
		return
	}

	payload, err := s.service.BattleFieldNewsPage(r.Context(), uid, battlefieldID, unionID, page, pageSize)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleTroopRoute(w http.ResponseWriter, r *http.Request) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	trimmed := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/troops/"), "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] != "callback" {
		http.NotFound(w, r)
		return
	}

	troopID, err := strconv.Atoi(parts[0])
	if err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid troop id")
		return
	}

	payload, err := s.service.CallbackTroop(r.Context(), uid, troopID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleMail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	folder := defaultString(r.URL.Query().Get("folder"), "inbox")
	page, err := strconv.Atoi(defaultString(r.URL.Query().Get("page"), "0"))
	if err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid mail page")
		return
	}

	payload, err := s.service.MailPage(r.Context(), uid, folder, page)
	if err != nil {
		httpjson.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleMailDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		Folder string `json:"folder"`
		IDs    []int  `json:"ids"`
		Page   int    `json:"page"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	page, err := s.service.DeleteMail(r.Context(), uid, payload.Folder, payload.IDs, payload.Page)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, page)
}

func (s Server) handleMailSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		ToName  string `json:"toName"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	detail, err := s.service.SendMail(r.Context(), uid, payload.ToName, payload.Title, payload.Content)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, detail)
}

func (s Server) handleMailRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpjson.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	trimmed := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/mail/"), "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid mail id")
		return
	}

	payload, err := s.service.MailDetail(r.Context(), uid, parts[0], id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusNotFound
		}
		httpjson.Error(w, status, err.Error())
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleReports(w http.ResponseWriter, r *http.Request) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	filter := defaultString(r.URL.Query().Get("filter"), "unread")
	page, err := strconv.Atoi(defaultString(r.URL.Query().Get("page"), "0"))
	if err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid report page")
		return
	}

	payload, err := s.service.Reports(r.Context(), uid, filter, page)
	if err != nil {
		httpjson.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleReportRoute(w http.ResponseWriter, r *http.Request) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	idText := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/reports/"), "/")
	if idText == "" {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.Atoi(idText)
	if err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid report id")
		return
	}

	payload, err := s.service.ReportDetail(r.Context(), uid, id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusNotFound
		}
		httpjson.Error(w, status, err.Error())
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleRankings(w http.ResponseWriter, r *http.Request) {
	kind := defaultString(r.URL.Query().Get("kind"), "user")
	page, err := strconv.Atoi(defaultString(r.URL.Query().Get("page"), "0"))
	if err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid ranking page")
		return
	}

	payload, err := s.service.Ranking(r.Context(), kind, page)
	if err != nil {
		httpjson.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleCities(w http.ResponseWriter, r *http.Request) {
	limit := 24
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	items, err := s.service.Cities(r.Context(), limit)
	if err != nil {
		httpjson.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpjson.Write(w, http.StatusOK, map[string]any{
		"items": items,
	})
}

func (s Server) handleCityRoute(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/cities/")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	cid, err := strconv.Atoi(parts[0])
	if err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid city id")
		return
	}

	if len(parts) == 1 && r.Method == http.MethodGet {
		s.handleCityDetail(w, r, cid)
		return
	}

	if len(parts) == 2 && parts[1] == "heroes" && r.Method == http.MethodGet {
		s.handleCityHeroes(w, r, cid)
		return
	}

	if len(parts) == 3 && parts[1] == "troops" && parts[2] == "dispatch" && r.Method == http.MethodPost {
		s.handleCityTroopDispatch(w, r, cid)
		return
	}

	if len(parts) == 2 && parts[1] == "barracks" && r.Method == http.MethodGet {
		s.handleCityBarracksSnapshot(w, r, cid)
		return
	}

	if len(parts) == 4 && parts[1] == "barracks" && parts[2] == "draft" && parts[3] == "start" && r.Method == http.MethodPost {
		s.handleCitySoldierDraftStart(w, r, cid)
		return
	}

	if len(parts) == 4 && parts[1] == "barracks" && parts[2] == "draft" && parts[3] == "cancel" && r.Method == http.MethodPost {
		s.handleCitySoldierDraftCancel(w, r, cid)
		return
	}

	if len(parts) == 2 && parts[1] == "research" && r.Method == http.MethodGet {
		s.handleCityResearchSnapshot(w, r, cid)
		return
	}

	if len(parts) == 3 && parts[1] == "research" && parts[2] == "start" && r.Method == http.MethodPost {
		s.handleCityResearchStart(w, r, cid)
		return
	}

	if len(parts) == 3 && parts[1] == "research" && parts[2] == "cancel" && r.Method == http.MethodPost {
		s.handleCityResearchCancel(w, r, cid)
		return
	}

	if len(parts) == 3 && parts[1] == "heroes" && parts[2] == "recruit" && r.Method == http.MethodPost {
		s.handleCityHeroRecruit(w, r, cid)
		return
	}

	if len(parts) == 4 && parts[1] == "heroes" && parts[3] == "armor" && r.Method == http.MethodGet {
		hid, err := strconv.Atoi(parts[2])
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid hero id")
			return
		}

		s.handleHeroArmorSnapshot(w, r, cid, hid)
		return
	}

	if len(parts) == 4 && parts[1] == "heroes" && parts[3] == "points" && r.Method == http.MethodPatch {
		hid, err := strconv.Atoi(parts[2])
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid hero id")
			return
		}

		s.handleHeroPointUpdate(w, r, cid, hid)
		return
	}

	if len(parts) == 5 && parts[1] == "heroes" && parts[3] == "armor" && parts[4] == "equip" && r.Method == http.MethodPost {
		hid, err := strconv.Atoi(parts[2])
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid hero id")
			return
		}

		s.handleHeroArmorEquip(w, r, cid, hid)
		return
	}

	if len(parts) == 5 && parts[1] == "heroes" && parts[3] == "armor" && parts[4] == "offload" && r.Method == http.MethodPost {
		hid, err := strconv.Atoi(parts[2])
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid hero id")
			return
		}

		s.handleHeroArmorOffload(w, r, cid, hid)
		return
	}

	if len(parts) == 5 && parts[1] == "heroes" && parts[3] == "armor" && parts[4] == "repair-all" && r.Method == http.MethodPost {
		hid, err := strconv.Atoi(parts[2])
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid hero id")
			return
		}

		s.handleHeroArmorRepairAll(w, r, cid, hid)
		return
	}

	if len(parts) == 5 && parts[1] == "heroes" && parts[3] == "armor" && parts[4] == "repair" && r.Method == http.MethodPost {
		hid, err := strconv.Atoi(parts[2])
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid hero id")
			return
		}

		s.handleHeroArmorRepair(w, r, cid, hid)
		return
	}

	if len(parts) == 5 && parts[1] == "heroes" && parts[3] == "armor" && parts[4] == "renovate-all" && r.Method == http.MethodPost {
		hid, err := strconv.Atoi(parts[2])
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid hero id")
			return
		}

		s.handleHeroArmorRenovateAll(w, r, cid, hid)
		return
	}

	if len(parts) == 5 && parts[1] == "heroes" && parts[3] == "armor" && parts[4] == "renovate" && r.Method == http.MethodPost {
		hid, err := strconv.Atoi(parts[2])
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid hero id")
			return
		}

		s.handleHeroArmorRenovate(w, r, cid, hid)
		return
	}

	if len(parts) == 5 && parts[1] == "heroes" && parts[3] == "armor" && parts[4] == "recycle" && r.Method == http.MethodPost {
		hid, err := strconv.Atoi(parts[2])
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid hero id")
			return
		}

		s.handleHeroArmorRecycle(w, r, cid, hid)
		return
	}

	if len(parts) == 2 && parts[1] == "chief" && r.Method == http.MethodPatch {
		s.handleCityChiefUpdate(w, r, cid)
		return
	}

	if len(parts) == 2 && parts[1] == "general" && r.Method == http.MethodPatch {
		s.handleCityGeneralUpdate(w, r, cid)
		return
	}

	if len(parts) == 2 && parts[1] == "counsellor" && r.Method == http.MethodPatch {
		s.handleCityCounsellorUpdate(w, r, cid)
		return
	}

	if len(parts) == 2 && parts[1] == "tax" && r.Method == http.MethodPatch {
		s.handleCityTaxUpdate(w, r, cid)
		return
	}

	if len(parts) == 2 && parts[1] == "production" && r.Method == http.MethodPatch {
		s.handleCityProductionUpdate(w, r, cid)
		return
	}

	if len(parts) == 3 && parts[1] == "buildings" && parts[2] == "options" && r.Method == http.MethodGet {
		s.handleCityBuildingOptions(w, r, cid)
		return
	}

	if len(parts) == 3 && parts[1] == "buildings" && parts[2] == "info" && r.Method == http.MethodGet {
		s.handleCityBuildingInfo(w, r, cid)
		return
	}

	if len(parts) == 3 && parts[1] == "buildings" && parts[2] == "create" && r.Method == http.MethodPost {
		s.handleCityBuildingCreate(w, r, cid)
		return
	}

	if len(parts) == 3 && parts[1] == "buildings" && parts[2] == "upgrade" && r.Method == http.MethodPost {
		s.handleCityBuildingUpgrade(w, r, cid)
		return
	}

	if len(parts) == 3 && parts[1] == "buildings" && parts[2] == "destroy" && r.Method == http.MethodPost {
		s.handleCityBuildingDestroy(w, r, cid)
		return
	}

	if len(parts) == 3 && parts[1] == "buildings" && parts[2] == "cancel" && r.Method == http.MethodPost {
		s.handleCityBuildingCancel(w, r, cid)
		return
	}

	if len(parts) == 4 && parts[1] == "buildings" && parts[2] == "speed-goods" && parts[3] == "use" && r.Method == http.MethodPost {
		s.handleCityBuildingSpeedGoodsUse(w, r, cid)
		return
	}

	if len(parts) == 3 && parts[1] == "buildings" && parts[2] == "speed-goods" && r.Method == http.MethodGet {
		s.handleCityBuildingSpeedGoods(w, r, cid)
		return
	}

	http.NotFound(w, r)
}

func (s Server) handleCityDetail(w http.ResponseWriter, r *http.Request, cid int) {
	detail, err := s.service.CityDetail(r.Context(), cid)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusNotFound
		}
		httpjson.Error(w, status, err.Error())
		return
	}

	httpjson.Write(w, http.StatusOK, detail)
}

func (s Server) handleCityHeroes(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	limit := 24
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}

	payload, err := s.service.CityHeroes(r.Context(), uid, cid, limit)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleCityChiefUpdate(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		HID int `json:"hid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	roster, err := s.service.UpdateCityChief(r.Context(), uid, cid, payload.HID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, roster)
}

func (s Server) handleCityGeneralUpdate(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		HID int `json:"hid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	roster, err := s.service.UpdateCityGeneral(r.Context(), uid, cid, payload.HID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, roster)
}

func (s Server) handleCityCounsellorUpdate(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		HID int `json:"hid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	roster, err := s.service.UpdateCityCounsellor(r.Context(), uid, cid, payload.HID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, roster)
}

func (s Server) handleHeroPointUpdate(w http.ResponseWriter, r *http.Request, cid int, hid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		Stat   string `json:"stat"`
		Amount int    `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	roster, err := s.service.AddHeroPoint(r.Context(), uid, cid, hid, payload.Stat, payload.Amount)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, roster)
}

func (s Server) handleHeroArmorSnapshot(w http.ResponseWriter, r *http.Request, cid int, hid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	snapshot, err := s.service.HeroArmorSnapshot(r.Context(), uid, cid, hid)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleHeroArmorEquip(w http.ResponseWriter, r *http.Request, cid int, hid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		SID   int `json:"sid"`
		Spart int `json:"spart"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.EquipHeroArmor(r.Context(), uid, cid, hid, payload.SID, payload.Spart)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleHeroArmorOffload(w http.ResponseWriter, r *http.Request, cid int, hid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		Spart int `json:"spart"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.OffloadHeroArmor(r.Context(), uid, cid, hid, payload.Spart)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleHeroArmorRepair(w http.ResponseWriter, r *http.Request, cid int, hid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		SID int `json:"sid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.RepairHeroArmor(r.Context(), uid, cid, hid, payload.SID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleHeroArmorRepairAll(w http.ResponseWriter, r *http.Request, cid int, hid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		SIDs []int `json:"sids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.RepairAllHeroArmor(r.Context(), uid, cid, hid, payload.SIDs)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleHeroArmorRenovate(w http.ResponseWriter, r *http.Request, cid int, hid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		SID int `json:"sid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.RenovateHeroArmor(r.Context(), uid, cid, hid, payload.SID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleHeroArmorRenovateAll(w http.ResponseWriter, r *http.Request, cid int, hid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		SIDs []int `json:"sids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.RenovateAllHeroArmor(r.Context(), uid, cid, hid, payload.SIDs)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleHeroArmorRecycle(w http.ResponseWriter, r *http.Request, cid int, hid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		SID int `json:"sid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.RecycleHeroArmor(r.Context(), uid, cid, hid, payload.SID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleCityHeroRecruit(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		RecruitID int `json:"recruitId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	roster, err := s.service.RecruitCityHero(r.Context(), uid, cid, payload.RecruitID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, roster)
}

func (s Server) handleCityTroopDispatch(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		TargetCID    int `json:"targetCid"`
		SoldierSID   int `json:"soldierSid"`
		SoldierCount int `json:"soldierCount"`
		Soldiers     []struct {
			SID   int   `json:"sid"`
			TID   int   `json:"tid"`
			Count int64 `json:"count"`
		} `json:"soldiers"`
		HeroID    int                   `json:"heroId"`
		HID       int                   `json:"hid"`
		Task      *int                  `json:"task"`
		Resources *legacy.TroopResource `json:"resources"`
		Resource  *legacy.TroopResource `json:"resource"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	task := 1
	if payload.Task != nil {
		task = *payload.Task
	}
	resource := legacy.TroopResource{}
	if payload.Resources != nil {
		resource = *payload.Resources
	} else if payload.Resource != nil {
		resource = *payload.Resource
	}
	soldiers := make(map[int]int64, len(payload.Soldiers)+1)
	for _, item := range payload.Soldiers {
		sid := item.SID
		if sid <= 0 {
			sid = item.TID
		}
		if sid > 0 && item.Count > 0 {
			soldiers[sid] += item.Count
		}
	}
	if len(soldiers) == 0 && payload.SoldierSID > 0 && payload.SoldierCount > 0 {
		soldiers[payload.SoldierSID] = int64(payload.SoldierCount)
	}

	heroID := payload.HeroID
	if heroID <= 0 {
		heroID = payload.HID
	}

	page, err := s.service.DispatchCityTroop(r.Context(), uid, cid, payload.TargetCID, soldiers, heroID, task, resource)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, page)
}

func (s Server) handleCityBarracksSnapshot(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	position, err := strconv.Atoi(r.URL.Query().Get("position"))
	if err != nil || position <= 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid barracks position")
		return
	}

	snapshot, err := s.service.CityBarracksSnapshot(r.Context(), uid, cid, position)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleCitySoldierDraftStart(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		Position int `json:"position"`
		SID      int `json:"sid"`
		Count    int `json:"count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.StartCitySoldierDraft(r.Context(), uid, cid, payload.Position, payload.SID, payload.Count)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleCitySoldierDraftCancel(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		Position int `json:"position"`
		QueueID  int `json:"queueId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.CancelCitySoldierDraft(r.Context(), uid, cid, payload.Position, payload.QueueID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleCityResearchSnapshot(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	position, err := strconv.Atoi(r.URL.Query().Get("position"))
	if err != nil || position <= 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid college position")
		return
	}

	snapshot, err := s.service.CityResearchSnapshot(r.Context(), uid, cid, position)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleCityResearchStart(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		Position int `json:"position"`
		TID      int `json:"tid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.StartCityResearch(r.Context(), uid, cid, payload.Position, payload.TID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleCityResearchCancel(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		Position int `json:"position"`
		TID      int `json:"tid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	snapshot, err := s.service.CancelCityResearch(r.Context(), uid, cid, payload.Position, payload.TID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, snapshot)
}

func (s Server) handleCityTaxUpdate(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		Tax int `json:"tax"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	detail, err := s.service.UpdateCityTax(r.Context(), uid, cid, payload.Tax)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, detail)
}

func (s Server) handleCityProductionUpdate(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload legacy.ProductionSettings
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	detail, err := s.service.UpdateCityProduction(r.Context(), uid, cid, payload)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, detail)
}

func (s Server) handleCityBuildingUpgrade(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		Position int `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	detail, err := s.service.UpgradeCityBuilding(r.Context(), uid, cid, payload.Position)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, detail)
}

func (s Server) handleCityBuildingDestroy(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		Position int `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	detail, err := s.service.DestroyCityBuilding(r.Context(), uid, cid, payload.Position)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, detail)
}

func (s Server) handleCityBuildingCancel(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		Position int `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	detail, err := s.service.CancelCityBuildingAction(r.Context(), uid, cid, payload.Position)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, detail)
}

func (s Server) handleCityBuildingOptions(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	position, err := strconv.Atoi(defaultString(r.URL.Query().Get("position"), "0"))
	if err != nil || position <= 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid building position")
		return
	}

	payload, err := s.service.CityBuildingPlacementOptions(r.Context(), uid, cid, position)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleCityBuildingInfo(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	position, err := strconv.Atoi(defaultString(r.URL.Query().Get("position"), "0"))
	if err != nil || position <= 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid building position")
		return
	}

	payload, err := s.service.CityBuildingInfo(r.Context(), uid, cid, position)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleCityBuildingCreate(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		Position int `json:"position"`
		BID      int `json:"bid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	detail, err := s.service.CreateCityBuilding(r.Context(), uid, cid, payload.Position, payload.BID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, detail)
}

func (s Server) handleCityBuildingSpeedGoods(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	position, err := strconv.Atoi(defaultString(r.URL.Query().Get("position"), "0"))
	if err != nil || position <= 0 {
		httpjson.Error(w, http.StatusBadRequest, "invalid building position")
		return
	}

	payload, err := s.service.BuildingSpeedGoods(r.Context(), uid, cid, position)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, payload)
}

func (s Server) handleCityBuildingSpeedGoodsUse(w http.ResponseWriter, r *http.Request, cid int) {
	uid, ok := s.requireAuth(w, r)
	if !ok {
		return
	}

	var payload struct {
		Position int `json:"position"`
		GID      int `json:"gid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	detail, err := s.service.UseBuildingSpeedGoods(r.Context(), uid, cid, payload.Position, payload.GID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}

	httpjson.Write(w, http.StatusOK, detail)
}

func (s Server) handleWorldMap(w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.Atoi(defaultString(r.URL.Query().Get("cid"), "215265"))
	if err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid cid")
		return
	}

	radius, err := strconv.Atoi(defaultString(r.URL.Query().Get("radius"), "8"))
	if err != nil {
		httpjson.Error(w, http.StatusBadRequest, "invalid radius")
		return
	}

	rawX := strings.TrimSpace(r.URL.Query().Get("x"))
	rawY := strings.TrimSpace(r.URL.Query().Get("y"))

	var worldMap legacy.WorldMap
	if rawX != "" && rawY != "" {
		focusX, err := strconv.Atoi(rawX)
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid x")
			return
		}

		focusY, err := strconv.Atoi(rawY)
		if err != nil {
			httpjson.Error(w, http.StatusBadRequest, "invalid y")
			return
		}

		worldMap, err = s.service.WorldMapAt(r.Context(), cid, focusX, focusY, radius)
	} else {
		worldMap, err = s.service.WorldMap(r.Context(), cid, radius)
	}
	if err != nil {
		httpjson.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpjson.Write(w, http.StatusOK, worldMap)
}

func (s Server) handleFrontend(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}

	indexPath := filepath.Join(s.cfg.FrontendDistDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <title>RXSG Refactor</title>
    <style>
      body { font-family: Segoe UI, sans-serif; background:#111827; color:#f8fafc; padding:48px; }
      .card { max-width:720px; border:1px solid #334155; background:#0f172a; border-radius:18px; padding:28px; }
      code { background:#1e293b; border-radius:6px; padding:2px 6px; }
    </style>
  </head>
  <body>
    <div class="card">
      <h1>Frontend assets are not built yet</h1>
      <p>Run <code>cd frontend</code>, <code>npm install</code> and <code>npm run build</code>, then refresh this page.</p>
      <p>The API is already available at <code>/api/health</code>.</p>
    </div>
  </body>
</html>`))
		return
	}

	target := filepath.Join(s.cfg.FrontendDistDir, strings.TrimPrefix(filepath.Clean(r.URL.Path), string(filepath.Separator)))
	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		http.ServeFile(w, r, target)
		return
	}

	http.ServeFile(w, r, indexPath)
}

func (s Server) currentUID(r *http.Request) (int, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return 0, false
	}

	record, ok := s.sessions.get(cookie.Value)
	if !ok {
		return 0, false
	}

	return record.UID, true
}

func (s Server) requireAuth(w http.ResponseWriter, r *http.Request) (int, bool) {
	uid, ok := s.currentUID(r)
	if !ok {
		httpjson.Error(w, http.StatusUnauthorized, "login required")
		return 0, false
	}
	return uid, true
}

func (s Server) writeSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s Server) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s Server) writeDomainError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, sql.ErrNoRows):
		status = http.StatusNotFound
	case errors.Is(err, legacy.ErrInvalid):
		status = http.StatusBadRequest
	case errors.Is(err, legacy.ErrForbidden):
		status = http.StatusForbidden
	}
	httpjson.Error(w, status, err.Error())
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withAccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(started))
	})
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func requestIPInt(r *http.Request) int64 {
	source := r.Header.Get("X-Forwarded-For")
	if source == "" {
		source = r.RemoteAddr
	}

	host := source
	if strings.Contains(host, ",") {
		host = strings.TrimSpace(strings.Split(host, ",")[0])
	}
	if strings.Contains(host, ":") {
		if parsedHost, _, err := net.SplitHostPort(host); err == nil {
			host = parsedHost
		}
	}

	ip := net.ParseIP(host).To4()
	if ip == nil {
		return 0
	}

	return int64(ip[3])<<24 + int64(ip[2])<<16 + int64(ip[1])<<8 + int64(ip[0])
}
