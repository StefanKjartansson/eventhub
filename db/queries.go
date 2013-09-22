package db

const (
	insertSQL = `
    INSERT INTO "event"
        (
            "key",
            "created",
            "updated",
            "payload",
            "description",
            "importance",
            "origin",
            "entities",
            "other_references",
            "actors",
            "tags"
        )
    VALUES
        (
            $1,
            now(),
            now(),
            $2,
            $3,
            $4,
            $5,
            ARRAY[$6],
            ARRAY[$7],
            ARRAY[$8],
            ARRAY[$9]
        )
    RETURNING
        "id",
        "created",
        "updated";
    `

	updateSQL = `
    UPDATE "event"
    SET
        "key" = $1,
        "payload" = $2,
        "description" = $3,
        "importance" = $4,
        "origin" = $5,
        "entities" = ARRAY[$6],
        "other_references" = ARRAY[$7],
        "actors" = ARRAY[$8],
        "tags" = ARRAY[$9],
        "updated" = now()
    WHERE
        "id" = $10
    RETURNING
        "updated";
    `

	selectByIdSQL = `
    SELECT
        *
    FROM
        "event"
    WHERE "id" = $1
    `
)
