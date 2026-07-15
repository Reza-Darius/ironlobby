-- name: GetCountryID :one
SELECT id FROM countries
WHERE country_tag = $1;

-- name: GetCountryTag :one
SELECT country_tag FROM countries
WHERE id = $1;

-- name: GetCountries :many
SELECT
    country_tag,
    country_name
FROM countries;
