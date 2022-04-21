package sql

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/isbn"
	"github.com/knakk/sirkulator/marc"
)

func readResource(res *sirkulator.Resource, t sirkulator.ResourceType) func(stmt *sqlite.Stmt) error {
	return func(stmt *sqlite.Stmt) error {
		res.Type = t
		res.ID = stmt.ColumnText(0)
		res.Label = stmt.ColumnText(1)
		if err := json.Unmarshal([]byte(stmt.ColumnText(2)), res.Data); err != nil {
			return err
		}
		res.CreatedAt = time.Unix(stmt.ColumnInt64(3), 0)
		res.UpdatedAt = time.Unix(stmt.ColumnInt64(4), 0)
		if n := stmt.ColumnInt64(5); n != 0 {
			res.ArchivedAt = time.Unix(n, 0)
		}
		return nil
	}
}

func readData(res *sirkulator.Resource, t sirkulator.ResourceType) func(stmt *sqlite.Stmt) error {
	// TODO try rewrite with generics
	switch t {
	case sirkulator.TypePerson:
		res.Data = &sirkulator.Person{}
		return readResource(res, t)
	case sirkulator.TypePublication:
		res.Data = &sirkulator.Publication{}
		return readResource(res, t)
	case sirkulator.TypeCorporation:
		res.Data = &sirkulator.Corporation{}
		return readResource(res, t)
	case sirkulator.TypeDewey:
		res.Data = &sirkulator.Dewey{}
		return readResource(res, t)
	case sirkulator.TypePublisher:
		res.Data = &sirkulator.Publisher{}
		return readResource(res, t)
	default:
		panic("sql.GetResource: readData: TODO")
	}
}

func readLinks(res *sirkulator.Resource) func(stmt *sqlite.Stmt) error {
	return func(stmt *sqlite.Stmt) error {
		k := stmt.ColumnText(0)
		v := stmt.ColumnText(1)
		if k == "isbn" {
			// TODO move this out closer to presentation layer
			v = isbn.Prettify(v)
		}
		res.Links = append(res.Links, [2]string{k, v})
		return nil
	}
}

func GetResource(conn *sqlite.Conn, t sirkulator.ResourceType, id string) (sirkulator.Resource, error) {
	var res sirkulator.Resource

	const qResouce = "SELECT id, label, data, created_at, updated_at, archived_at FROM resource WHERE type=? AND id=?"
	if err := sqlitex.Exec(conn, qResouce, readData(&res, t), t.String(), id); err != nil {
		return res, fmt.Errorf("sql.GetResource(%s, %s): %w", t.String(), id, err)
	}
	if res.ID == "" {
		return res, sirkulator.ErrNotFound
	}

	const qLinks = "SELECT type, id FROM link WHERE resource_id=? ORDER BY type"
	if err := sqlitex.Exec(conn, qLinks, readLinks(&res), id); err != nil {
		return res, fmt.Errorf("sql.GetResource(%s, %s): %w", t.String(), id, err)
	}

	return res, nil
}

func GetDeweyParts(conn *sqlite.Conn, id string) ([][2]string, error) {
	const q = `
      SELECT res.id, res.label
      FROM relation rel
      JOIN resource res ON (rel.to_id=res.id AND rel.type='has_part')
     WHERE rel.from_id=?` // TODO consider json_extract(data, '$.name') as label

	var res [][2]string
	fn := func(stmt *sqlite.Stmt) error {
		res = append(res, [2]string{stmt.ColumnText(0), stmt.ColumnText(1)})
		return nil
	}
	if err := sqlitex.Exec(conn, q, fn, id); err != nil {
		return res, fmt.Errorf("sql.GetDeweyParts(%q): %w", id, err)
	}

	return res, nil
}

func GetDeweyPartsOf(conn *sqlite.Conn, id, from, to string, limit int) ([][2]string, bool, error) {
	var (
		q        string
		fromOrTo string
	)
	if from != "" || to == "" {
		q = `
        SELECT res.id, res.label
          FROM relation rel
          JOIN resource res ON (rel.from_id=res.id AND rel.type='has_part')
         WHERE rel.to_id=?
           AND res.id > ?
         ORDER BY res.id ASC
         LIMIT ?` // TODO consider json_extract(data, '$.name') as label
		fromOrTo = from
	} else {
		q = `
        SELECT res.id, res.label
          FROM relation rel
          JOIN resource res ON (rel.from_id=res.id AND rel.type='has_part')
         WHERE rel.to_id=?
           AND res.id < ?
         ORDER BY res.id DESC
         LIMIT ?` // TODO consider json_extract(data, '$.name') as label
		fromOrTo = to
	}

	var res [][2]string
	hasMore := false
	fn := func(stmt *sqlite.Stmt) error {
		res = append(res, [2]string{stmt.ColumnText(0), stmt.ColumnText(1)})
		return nil
	}
	if err := sqlitex.Exec(conn, q, fn, id, fromOrTo, limit+1); err != nil {
		return res, hasMore, fmt.Errorf("sql.GetDeweyPartsOf(%q, %s, %s, %d): %w", id, from, to, limit, err)
	}

	// We try to fetch one more that requested, so that we know if there
	// are more results to be had.
	if len(res) > limit {
		hasMore = true
		res = res[:limit]
	}

	if to != "" {
		// reverse res
		for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
			res[i], res[j] = res[j], res[i]
		}
	}

	return res, hasMore, nil
}

func GetDeweyPartsOfCount(conn *sqlite.Conn, id string) (int, error) {
	stmt := conn.Prep(`
        SELECT count(res.id), res.label
          FROM relation rel
          JOIN resource res ON (rel.from_id=res.id AND rel.type='has_part')
         WHERE rel.to_id=$id`)
	stmt.SetText("$id", id)

	n, err := sqlitex.ResultInt(stmt)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func GetDeweyPublicationsCount(conn *sqlite.Conn, id string) (int, error) {
	var stmt *sqlite.Stmt
	if strings.HasPrefix(id, "T") {
		stmt = conn.Prep(`
		SELECT count(DISTINCT res.id)
		  FROM resource res
		  JOIN relation rel ON (rel.from_id=res.id AND rel.type='has_classification')
		  JOIN relation prel ON (rel.to_id=prel.from_id AND prel.type='has_part')
		 WHERE prel.to_id=$id AND res.type='publication'`)
	} else {
		stmt = conn.Prep(`
		SELECT count(DISTINCT res.id)
		  FROM resource res
		  JOIN relation rel ON (rel.from_id=res.id AND rel.type='has_classification')
		 WHERE rel.to_id=$id AND res.type='publication'`)
	}

	stmt.SetText("$id", id)
	n, err := sqlitex.ResultInt(stmt)
	if err != nil {
		return 0, fmt.Errorf("sql.GetDeweyPublicationsCount(%q): %w", id, err)
	}
	return n, nil
}

func GetDeweySubPublicationsCount(conn *sqlite.Conn, id string) (int, error) {
	var stmt *sqlite.Stmt
	if strings.HasPrefix(id, "T") {
		stmt = conn.Prep(`
		WITH RECURSIVE dewey (id) AS (
			SELECT $id AS id

			UNION ALL

			SELECT from_id FROM relation
			JOIN dewey ON to_id=dewey.id
			WHERE type='has_parent'
		)
	SELECT count(DISTINCT res.id)
	  FROM resource res
	  JOIN relation rel ON (rel.from_id=res.id AND rel.type='has_classification')
	  JOIN relation prel ON (rel.to_id=prel.from_id AND prel.type='has_part')
	  JOIN dewey ON (prel.to_id=dewey.id)
	 WHERE res.type='publication'`)
	} else {
		stmt = conn.Prep(`
		WITH RECURSIVE dewey (id) AS (
			SELECT $id AS id

			UNION ALL

			SELECT from_id FROM relation
			JOIN dewey ON to_id=dewey.id
			WHERE type='has_parent'
		)
		SELECT count(DISTINCT res.id)
		 FROM resource res
		 JOIN relation rel ON (rel.from_id=res.id AND rel.type='has_classification')
		 JOIN dewey ON (rel.to_id=dewey.id)
		WHERE res.type='publication'`)
	}

	stmt.SetText("$id", id)
	n, err := sqlitex.ResultInt(stmt)
	if err != nil {
		return 0, fmt.Errorf("sql.GetDeweySubPublicationsCount(%q): %w", id, err)
	}
	return n, nil
}

func ltOrGt(dir string) string {
	if strings.ToLower(dir) == "desc" {
		return ">"
	}
	return "<"
}

func reverseDir(dir string) string {
	if strings.ToLower(dir) == "desc" {
		return "asc"
	}
	return "desc"
}

func deweyPublicationsQuery(id string, params DeweyPublicationsParams) string {
	if params.SortBy == "dewey" {
		params.SortBy = "rel.to_id"
	}

	order := params.SortDir
	dir := ltOrGt(order)
	if order == "asc" && params.From == "" {
		dir = ">"
	}
	if params.To != "" {
		if order == "asc" {
			dir = "<"
		}
		order = reverseDir(order)
	}
	if params.From != "" {
		dir = ltOrGt(reverseDir(order))
	}
	if strings.HasPrefix(id, "T") {
		if params.InclSub {
			return fmt.Sprintf(`
			WITH RECURSIVE dewey (id) AS (
				SELECT ? AS id

				UNION ALL

				SELECT from_id FROM relation
				JOIN dewey ON to_id=dewey.id
				WHERE type='has_parent'
			)
			SELECT res.id, res.label, json_extract(res.data, '$.year') AS year, group_concat(rel.to_id, ', ') AS deweynumbers
		  FROM resource res
		  JOIN relation rel ON (rel.from_id=res.id AND rel.type='has_classification')
		  JOIN relation prel ON (rel.to_id=prel.from_id AND prel.type='has_part')
		  JOIN dewey ON (prel.to_id=dewey.id)
		 WHERE res.type='publication' AND
			   %s %s ? OR (%s = ? AND res.id %s ?)
	  GROUP BY res.id
	  ORDER BY %s %s, res.id %s
	  LIMIT ?`, params.SortBy, dir, params.SortBy, dir, params.SortBy, order, order)
		}

		return fmt.Sprintf(`
		SELECT res.id, res.label, json_extract(res.data, '$.year') AS year, group_concat(rel.to_id, ', ') AS deweynumbers
		  FROM resource res
		  JOIN relation rel ON (rel.from_id=res.id AND rel.type='has_classification')
		  JOIN relation prel ON (rel.to_id=prel.from_id AND prel.type='has_part')
		 WHERE prel.to_id=? AND
		       res.type='publication' AND
			   %s %s ? OR (%s = ? AND res.id %s ?)
	  GROUP BY res.id
	  ORDER BY %s %s, res.id %s
	  LIMIT ?`, params.SortBy, dir, params.SortBy, dir, params.SortBy, order, order)
	}

	if params.InclSub {
		return fmt.Sprintf(`WITH RECURSIVE dewey (id) AS (
			SELECT ? AS id

			UNION ALL

			SELECT from_id FROM relation
			JOIN dewey ON to_id=dewey.id
			WHERE type='has_parent'
		)
		SELECT res.id, res.label, json_extract(res.data, '$.year') AS year, group_concat(rel.to_id, ', ') AS deweynumbers
		 FROM resource res
		 JOIN relation rel ON (rel.from_id=res.id AND rel.type='has_classification')
		 JOIN dewey ON (rel.to_id=dewey.id)
		WHERE res.type='publication' AND
			  %s %s ? OR (%s = ? AND res.id %s ?)
	 GROUP BY res.id
	 ORDER BY %s %s, res.id %s
	 LIMIT ?`, params.SortBy, dir, params.SortBy, dir, params.SortBy, order, order)
	}

	return fmt.Sprintf(`
    SELECT res.id, res.label, json_extract(res.data, '$.year') AS year, group_concat(rel.to_id, ', ') AS deweynumbers
      FROM resource res
      JOIN relation rel ON (rel.from_id=res.id AND rel.type='has_classification')
     WHERE rel.to_id=? AND
           res.type='publication' AND
		   %s %s ? OR (%s = ? AND res.id %s ?)
  GROUP BY res.id
  ORDER BY %s %s, res.id %s
  LIMIT ?`, params.SortBy, dir, params.SortBy, dir, params.SortBy, order, order)

}

type DeweyPublicationsParams struct {
	InclSub bool
	From    string
	FromID  string
	To      string
	ToID    string
	SortBy  string
	SortDir string
	Limit   int
}

// [4]string{id, label, year, dewey}
func GetDeweyPublications(conn *sqlite.Conn, id string, params DeweyPublicationsParams) ([][4]string, bool, error) {
	q := deweyPublicationsQuery(id, params)

	var res [][4]string
	hasMore := false
	fn := func(stmt *sqlite.Stmt) error {
		res = append(res, [4]string{
			stmt.ColumnText(0),
			stmt.ColumnText(1),
			stmt.ColumnText(2),
			stmt.ColumnText(3)})
		return nil
	}
	var fromOrTo any
	fromOrTo = params.To
	fromOrToID := params.ToID
	if params.From != "" || params.To == "" {
		fromOrTo = params.From
		fromOrToID = params.FromID
	}
	if params.SortBy == "year" {
		n, _ := strconv.Atoi(fromOrTo.(string))
		fromOrTo = n
	}

	if err := sqlitex.Exec(conn, q, fn, id, fromOrTo, fromOrTo, fromOrToID, params.Limit+1); err != nil {
		return res, hasMore, fmt.Errorf("sql.GetDeweyPublications(%q): %w", id, err)
	}

	// We try to fetch one more that requested, so that we know if there
	// are more results to be had.
	if len(res) > params.Limit {
		hasMore = true
		res = res[:params.Limit]
	}

	if params.To != "" {
		// reverse res
		for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
			res[i], res[j] = res[j], res[i]
		}
	}

	return res, hasMore, nil
}

func GetDeweyParents(conn *sqlite.Conn, id string) ([][2]string, error) {
	const q = `
    WITH RECURSIVE parents (id, label) AS (
        SELECT res.id, res.label
          FROM relation rel
          JOIN resource res ON (rel.to_id=res.id AND rel.type='has_parent')
         WHERE rel.from_id=?

        UNION ALL

        SELECT res.id, res.label
          FROM relation rel
          JOIN resource res ON (rel.to_id=res.id AND rel.type='has_parent')
          JOIN parents p ON (p.id = rel.from_id)
    )
    SELECT * FROM parents;`
	// TODO consider json_extract(data, '$.name') as label

	var res [][2]string
	fn := func(stmt *sqlite.Stmt) error {
		res = append(res, [2]string{stmt.ColumnText(0), stmt.ColumnText(1)})
		return nil
	}
	if err := sqlitex.Exec(conn, q, fn, id); err != nil {
		return res, fmt.Errorf("sql.GetDeweyParents(%q): %w", id, err)
	}

	return res, nil
}

func GetDeweyChildren(conn *sqlite.Conn, id string) ([][2]string, error) {
	const q = `
    SELECT res.id, res.label
      FROM relation rel
      JOIN resource res ON (rel.from_id=res.id AND rel.type='has_parent')
     WHERE rel.to_id=?` // TODO consider json_extract(data, '$.name') as label

	var res [][2]string
	fn := func(stmt *sqlite.Stmt) error {
		res = append(res, [2]string{stmt.ColumnText(0), stmt.ColumnText(1)})
		return nil
	}
	if err := sqlitex.Exec(conn, q, fn, id); err != nil {
		return res, fmt.Errorf("sql.GetDeweyChildren(%q): %w", id, err)
	}

	return res, nil
}

// TODO orderBy year|label orderAsc bool
func GetAgentContributions(conn *sqlite.Conn, id string, sortBy string, sortAsc bool) ([]sirkulator.AgentContribution, error) {
	var res []sirkulator.AgentContribution

	sortDir := "DESC"
	if sortAsc {
		sortDir = "ASC"
	}
	switch sortBy {
	case "year":
	// OK
	case "label":
		sortBy = "res_label"
	default:
		sortBy = "year"
	}

	q := fmt.Sprintf(`
    SELECT
        r.from_id,
        resource.type as res_type,
        resource.label as res_label,
        json_extract(resource.data, '$.year') AS year,
        GROUP_CONCAT(json_extract(r.data, '$.role')) as role
    FROM
        relation r
        JOIN resource ON (from_id=resource.id)
    WHERE
        r.type='has_contributor'
    AND to_id=?
    GROUP BY r.from_id
    ORDER BY %s %s`, sortBy, sortDir)

	fn := func(stmt *sqlite.Stmt) error {
		c := sirkulator.AgentContribution{}
		c.ID = stmt.ColumnText(0)
		c.Type = sirkulator.ParseResourceType(stmt.ColumnText(1))
		c.Label = stmt.ColumnText(2)
		c.Year = stmt.ColumnInt(3)
		for _, role := range strings.Split(stmt.ColumnText(4), ",") {
			rel, err := marc.ParseRelator(role)
			if err == nil {
				c.Roles = append(c.Roles, rel)
			}
		}
		res = append(res, c)
		return nil
	}

	if err := sqlitex.Exec(conn, q, fn, id); err != nil {
		return res, fmt.Errorf("sql.GetAgentContributions(%q): %w", id, err)
	}

	return res, nil
}

func GetPublisherPublications(conn *sqlite.Conn, id string, sortBy string, sortAsc bool) ([]sirkulator.PublisherPublication, error) {
	var res []sirkulator.PublisherPublication

	sortDir := "DESC"
	if sortAsc {
		sortDir = "ASC"
	}
	switch sortBy {
	case "year":
	// OK
	case "label":
		sortBy = "res_label"
	default:
		sortBy = "year"
	}

	q := fmt.Sprintf(`
    SELECT
        r.from_id,
        resource.label as res_label,
        json_extract(resource.data, '$.year') AS year
    FROM
        relation r
        JOIN resource ON (from_id=resource.id)
    WHERE
        r.type='published_by'
    AND to_id=?
    ORDER BY %s %s`, sortBy, sortDir)

	fn := func(stmt *sqlite.Stmt) error {
		r := sirkulator.PublisherPublication{}
		r.ID = stmt.ColumnText(0)
		r.Type = sirkulator.TypePublication
		r.Label = stmt.ColumnText(1)
		r.Year = stmt.ColumnInt(2)
		res = append(res, r)
		return nil
	}

	if err := sqlitex.Exec(conn, q, fn, id); err != nil {
		return res, fmt.Errorf("sql.GetPublisherPublications(%q): %w", id, err)
	}

	return res, nil
}

func GetPublcationContributors(conn *sqlite.Conn, id string) ([]sirkulator.PublicationContribution, error) {
	var res []sirkulator.PublicationContribution

	const q = `
    SELECT
        r.to_id,
        resource.type as res_type,
        resource.label as res_label,
        GROUP_CONCAT(json_extract(r.data, '$.role')) as role
    FROM
        relation r
        JOIN resource ON (to_id=resource.id)
    WHERE
        r.type='has_contributor'
    AND from_id=?
    GROUP BY r.to_id`

	fn := func(stmt *sqlite.Stmt) error {
		c := sirkulator.PublicationContribution{}
		c.Agent.ID = stmt.ColumnText(0)
		c.Agent.Type = sirkulator.ParseResourceType(stmt.ColumnText(1))
		c.Agent.Label = stmt.ColumnText(2)
		for _, role := range strings.Split(stmt.ColumnText(3), ",") {
			rel, err := marc.ParseRelator(role)
			if err == nil {
				c.Roles = append(c.Roles, rel)
			}
		}
		res = append(res, c)
		return nil
	}

	if err := sqlitex.Exec(conn, q, fn, id); err != nil {
		return res, fmt.Errorf("sql.GetPublicationContributions(%q): %w", id, err)
	}

	return res, nil
}

func GetPublcationRelations(conn *sqlite.Conn, id string) ([]sirkulator.RelationExp, error) {
	var res []sirkulator.RelationExp

	const q = `
    SELECT
        rel.id,
        rel.type,
        rel.to_id,
        rel.data,
        res.type AS res_type,
        IFNULL(json_extract(rel.data, '$.label'), res.label) AS res_label
    FROM
        relation rel
        LEFT JOIN resource res ON (rel.to_id=res.id)
    WHERE
        rel.from_id=?`

	fn := func(stmt *sqlite.Stmt) error {
		r := sirkulator.RelationExp{
			Relation: sirkulator.Relation{
				ID:   stmt.ColumnInt64(0),
				Type: stmt.ColumnText(1),
				ToID: stmt.ColumnText(2),
			},
			To: sirkulator.SimpleResource{
				ID:    stmt.ColumnText(2),
				Type:  sirkulator.ParseResourceType(stmt.ColumnText(4)),
				Label: stmt.ColumnText(5),
			},
		}
		var data map[string]any
		if err := json.Unmarshal([]byte(stmt.ColumnText(3)), &data); err != nil {
			return err // TODO annotate
		}
		r.Relation.Data = data

		res = append(res, r)
		return nil
	}

	if err := sqlitex.Exec(conn, q, fn, id); err != nil {
		return res, fmt.Errorf("sql.GetPublicationRelations(%q): %w", id, err)
	}

	return res, nil
}

func GetPublcationReviews(conn *sqlite.Conn, id string) ([]sirkulator.Relation, error) {
	var res []sirkulator.Relation

	const q = `
    SELECT
        type,
        data
    FROM
        relation
    WHERE
        from_id=?
    AND to_id IS NULL
    ORDER BY queued_at`

	fn := func(stmt *sqlite.Stmt) error {
		rel := sirkulator.Relation{
			FromID: id,
			Type:   stmt.ColumnText(0),
		}
		var data map[string]any
		if err := json.Unmarshal([]byte(stmt.ColumnText(1)), &data); err != nil {
			return err // TODO annotate
		}
		rel.Data = data
		if data == nil {
			return errors.New("review has no data")
		}
		res = append(res, rel)
		return nil
	}

	if err := sqlitex.Exec(conn, q, fn, id); err != nil {
		return res, fmt.Errorf("sql.GetPublicationReviews(%q): %w", id, err)
	}

	return res, nil
}

// TODO pagination? offset by rowid
func GetAllReviews(conn *sqlite.Conn, limit int) ([]sirkulator.RelationExp, error) {
	var res []sirkulator.RelationExp

	const q = `
    SELECT
        rel.id,
        rel.from_id,
        rel.type,
        rel.data,
        res.type,
        res.label
    FROM
        relation rel
        JOIN resource res ON (rel.from_id=res.id)
    WHERE
        rel.to_id IS NULL
    ORDER BY rel.queued_at
    LIMIT ?`

	fn := func(stmt *sqlite.Stmt) error {
		rel := sirkulator.RelationExp{
			Relation: sirkulator.Relation{
				ID:     stmt.ColumnInt64(0),
				FromID: stmt.ColumnText(1),
				Type:   stmt.ColumnText(2),
			},
			From: sirkulator.SimpleResource{
				ID:    stmt.ColumnText(1),
				Type:  sirkulator.ParseResourceType(stmt.ColumnText(4)),
				Label: stmt.ColumnText(5),
			},
		}
		var data map[string]any
		if err := json.Unmarshal([]byte(stmt.ColumnText(3)), &data); err != nil {
			return err // TODO annotate
		}
		rel.Data = data
		if data == nil {
			return errors.New("review has no data")
		}
		res = append(res, rel)
		return nil
	}

	if err := sqlitex.Exec(conn, q, fn, limit); err != nil {
		return res, fmt.Errorf("sql.GetAllReviews(%d): %w", limit, err)
	}

	return res, nil
}

func GetImage(conn *sqlite.Conn, id string) (*sirkulator.Image, error) {
	var img sirkulator.Image
	const q = "SELECT type, width, height FROM files.image WHERE id=?"
	fn := func(stmt *sqlite.Stmt) error {
		img.ID = id
		img.Type = stmt.ColumnText(0)
		img.Width = stmt.ColumnInt(1)
		img.Height = stmt.ColumnInt(2)
		return nil
	}

	if err := sqlitex.Exec(conn, q, fn, id); err != nil {
		return nil, fmt.Errorf("sql.GetImage(%q): %w", id, err)
	}
	if img.ID == "" {
		return nil, sirkulator.ErrNotFound
	}

	return &img, nil
}

func UpdateResource(conn *sqlite.Conn, res sirkulator.Resource, label string) (err error) {
	defer sqlitex.Save(conn)(&err)
	stmt := conn.Prep(`
            UPDATE resource SET data=$data, label=$label, updated_at=$updated_at
            WHERE id=$id
        `)

	stmt.SetText("$id", res.ID)
	stmt.SetInt64("$updated_at", time.Now().Unix())
	stmt.SetText("$label", label)
	b, err := json.Marshal(res.Data)
	if err != nil {
		return err // TODO annotate
	}
	stmt.SetBytes("$data", b)
	if _, err := stmt.Step(); err != nil {
		return err // TODO annotate
	}
	return nil
}

func GetResourceTexts(conn *sqlite.Conn, id string) ([]sirkulator.ResourceText, error) {
	var res []sirkulator.ResourceText

	const q = `
		SELECT
			id,
			text,
			format,
			source,
			source_url,
			updated_at
		FROM resource_text
		WHERE resource_id=?
	`

	fn := func(stmt *sqlite.Stmt) error {
		res = append(res, sirkulator.ResourceText{
			ID:        stmt.ColumnInt64(0),
			Text:      stmt.ColumnText(1),
			Format:    stmt.ColumnText(2),
			Source:    stmt.ColumnText(3),
			SourceURL: stmt.ColumnText(4),
			UpdatedAt: time.Unix(stmt.ColumnInt64(5), 0),
		})
		return nil
	}

	if err := sqlitex.Exec(conn, q, fn, id); err != nil {
		return res, fmt.Errorf("sql.GetResourceTexts(%s): %w", id, err)
	}

	return res, nil
}
