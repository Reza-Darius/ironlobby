-- name: GetCountry :one
SELECT id FROM countries WHERE country_tag = $1;

-- name: ListCountries :many
SELECT country_tag FROM countries;
