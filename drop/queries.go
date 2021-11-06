package drop

import sq "github.com/Masterminds/squirrel"

var qRandomDrop, _ = sq.Select(
	`drops.id AS "drops.id"`,
	`drops.title AS "drops.title"`,
	`drops.status AS "drops.status"`,
	`drops.moved_at AS "drops.moved_at"`,
	`drops.article_id AS "drops.article_id"`,
	`drops.user_id AS "drops.user_id"`,
	`drops.created_at AS "drops.created_at"`,
	`drops.updated_at AS "drops.updated_at"`,
	`articles.id AS "articles.id"`,
	`articles.title AS "articles.title"`,
	`articles.url AS "articles.url"`,
	`articles.created_at AS "articles.created_at"`,
	`articles.updated_at AS "articles.updated_at"`,
).From("drops").
	Join("articles ON articles.id = drops.article_id").
	Limit(1).
	OrderBy("random()").
	MustSql()
