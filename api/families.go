package api

import (
	"fmt"
	"net/http"
	"strings"
)

func (s *Server) familyRoutes() {
	s.memberRoute("GET /api/families", s.families)
	s.memberRoute("POST /api/families", s.createFamily)
	s.memberRoute("GET /api/families/{family}", s.familyDetails)
	s.memberRoute("DELETE /api/families/{family}", s.deleteFamily)
	s.memberRoute("POST /api/families/{family}/members", s.familyMemberAction)
	s.memberRoute("POST /api/families/{family}/people", s.familyPerson)
	s.memberRoute("DELETE /api/families/{family}/people/{person}", s.deleteFamilyPerson)
	s.memberRoute("POST /api/families/{family}/edges", s.familyEdge)
	s.memberRoute("DELETE /api/families/{family}/edges", s.deleteFamilyEdge)
}
func (s *Server) families(w http.ResponseWriter, r *http.Request, id string) {
	s.memberRows(w, r, `SELECT f.id,f.name,f.owner=$1 mine,COALESCE(m.accepted,false) accepted,a.handle owner_handle FROM family_groups f JOIN member_accounts a ON a.id=f.owner LEFT JOIN family_members m ON m.group_id=f.id AND m.account_id=$1 WHERE f.owner=$1 OR (m.account_id=$1 AND NOT member_blocked(f.owner,$1)) ORDER BY f.created_at`, id)
}
func (s *Server) createFamily(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Name string `json:"name"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if strings.TrimSpace(in.Name) == "" || len(in.Name) > 100 {
		problem(w, 400, fmt.Errorf("family name required (maximum 100 characters)"))
		return
	}
	s.memberExec(w, r, `INSERT INTO family_groups(id,owner,name) VALUES($1,$2,$3)`, randomToken(), id, in.Name)
}
func (s *Server) familyPermission(r *http.Request, id string, edit bool) bool {
	var ok bool
	err := s.membersDB().QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM family_groups f WHERE f.id=$1 AND (f.owner=$2 OR (family_access(f.id,$2) AND EXISTS(SELECT 1 FROM family_members m WHERE m.group_id=f.id AND m.account_id=$2 AND (NOT $3 OR m.role='editor')))))`, r.PathValue("family"), id, edit).Scan(&ok)
	return err == nil && ok
}
func (s *Server) familyDetails(w http.ResponseWriter, r *http.Request, id string) {
	if !s.familyPermission(r, id, false) {
		problem(w, 404, fmt.Errorf("family unavailable; accept your invitation first"))
		return
	}
	s.memberRows(w, r, `SELECT f.id,f.name,f.owner=$2 mine,
 COALESCE((SELECT json_agg(json_build_object('id',p.id,'name',p.name,'note',p.note) ORDER BY p.created_at) FROM family_people p WHERE p.group_id=f.id),'[]') people,
 COALESCE((SELECT json_agg(json_build_object('source',e.source,'target',e.target,'relation',e.relation)) FROM family_edges e WHERE e.group_id=f.id),'[]') edges,
 COALESCE((SELECT json_agg(json_build_object('id',a.id,'handle',a.handle,'accepted',m.accepted,'role',m.role)) FROM family_members m JOIN member_accounts a ON a.id=m.account_id WHERE m.group_id=f.id),'[]') members
 FROM family_groups f WHERE f.id=$1`, r.PathValue("family"), id)
}
func (s *Server) deleteFamily(w http.ResponseWriter, r *http.Request, id string) {
	s.memberExec(w, r, `DELETE FROM family_groups WHERE id=$1 AND owner=$2`, r.PathValue("family"), id)
}
func (s *Server) familyMemberAction(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Action string `json:"action"`
		Handle string `json:"handle"`
		Role   string `json:"role"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	g := r.PathValue("family")
	switch in.Action {
	case "accept":
		s.memberExec(w, r, `UPDATE family_members m SET accepted=true WHERE m.group_id=$1 AND m.account_id=$2 AND NOT member_blocked($2,(SELECT owner FROM family_groups WHERE id=$1))`, g, id)
	case "leave":
		s.memberExec(w, r, `DELETE FROM family_members WHERE group_id=$1 AND account_id=$2`, g, id)
	case "invite":
		if in.Role != "editor" {
			in.Role = "viewer"
		}
		s.memberExec(w, r, `INSERT INTO family_members(group_id,account_id,role) SELECT f.id,a.id,$4 FROM family_groups f JOIN member_accounts a ON a.handle=$3 WHERE f.id=$1 AND f.owner=$2 AND a.id<>$2 AND NOT member_blocked($2,a.id) ON CONFLICT(group_id,account_id) DO UPDATE SET role=EXCLUDED.role`, g, id, strings.ToLower(strings.TrimSpace(in.Handle)), in.Role)
	case "remove":
		s.memberExec(w, r, `DELETE FROM family_members m USING family_groups f,member_accounts a WHERE m.group_id=f.id AND m.account_id=a.id AND f.id=$1 AND f.owner=$2 AND a.handle=$3`, g, id, in.Handle)
	default:
		problem(w, 400, fmt.Errorf("invalid action"))
	}
}
func (s *Server) familyPerson(w http.ResponseWriter, r *http.Request, id string) {
	if !s.familyPermission(r, id, true) {
		problem(w, 403, fmt.Errorf("family editor permission required"))
		return
	}
	var in struct {
		Name    string `json:"name"`
		Note    string `json:"note"`
		Consent bool   `json:"consent"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if !in.Consent || strings.TrimSpace(in.Name) == "" || len(in.Name) > 100 || len(in.Note) > 300 {
		problem(w, 400, fmt.Errorf("name, sharing permission and short note required"))
		return
	}
	s.memberExec(w, r, `INSERT INTO family_people(id,group_id,name,note) VALUES($1,$2,$3,$4)`, randomToken(), r.PathValue("family"), in.Name, in.Note)
}
func (s *Server) deleteFamilyPerson(w http.ResponseWriter, r *http.Request, id string) {
	if !s.familyPermission(r, id, true) {
		problem(w, 403, fmt.Errorf("family editor permission required"))
		return
	}
	s.memberExec(w, r, `DELETE FROM family_people WHERE id=$1 AND group_id=$2`, r.PathValue("person"), r.PathValue("family"))
}

type familyEdgeInput struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	Relation string `json:"relation"`
}

func (s *Server) familyEdge(w http.ResponseWriter, r *http.Request, id string) {
	if !s.familyPermission(r, id, true) {
		problem(w, 403, fmt.Errorf("family editor permission required"))
		return
	}
	var in familyEdgeInput
	if !memberInput(w, r, &in) {
		return
	}
	s.memberExec(w, r, `INSERT INTO family_edges(group_id,source,target,relation) SELECT $1,p.id,q.id,$4 FROM family_people p JOIN family_people q ON q.group_id=p.group_id WHERE p.group_id=$1 AND p.id=$2 AND q.id=$3 ON CONFLICT(source,target,relation) DO UPDATE SET relation=EXCLUDED.relation`, r.PathValue("family"), in.Source, in.Target, in.Relation)
}
func (s *Server) deleteFamilyEdge(w http.ResponseWriter, r *http.Request, id string) {
	if !s.familyPermission(r, id, true) {
		problem(w, 403, fmt.Errorf("family editor permission required"))
		return
	}
	var in familyEdgeInput
	if !memberInput(w, r, &in) {
		return
	}
	s.memberExec(w, r, `DELETE FROM family_edges WHERE group_id=$1 AND source=$2 AND target=$3 AND relation=$4`, r.PathValue("family"), in.Source, in.Target, in.Relation)
}
