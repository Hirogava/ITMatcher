CREATE TABLE IF NOT EXISTS "hr" (
  "id" serial PRIMARY KEY,
  "email" text UNIQUE,
  "hash_password" text,
  "username" text UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS "vacancies" (
  "id" serial PRIMARY KEY,
  "name" text
);

CREATE TABLE IF NOT EXISTS "vacantion_hard_skills" (
  "id" serial PRIMARY KEY,
  "vacancy_id" integer,
  "hard_skill_id" integer
);

CREATE TABLE IF NOT EXISTS "vacantion_soft_skills" (
  "id" serial PRIMARY KEY,
  "vacancy_id" integer,
  "soft_skill_id" integer
);

CREATE TABLE IF NOT EXISTS "finders" (
  "id" serial PRIMARY KEY,
  "portfolio" boolean,
  "hr_id" integer
);

CREATE TABLE IF NOT EXISTS "resumes" (
  "id" serial PRIMARY KEY,
  "first_name" text,
  "last_name" text,
  "surname" text,
  "phone_number" text,
  "email" text,
  "vacancy_id" integer,
  "finder_id" integer
);

CREATE TABLE IF NOT EXISTS "portfolio" (
  "id" serial PRIMARY KEY,
  "finder_id" integer
);

CREATE TABLE IF NOT EXISTS "portfolio_hard_skills" (
  "id" serial PRIMARY KEY,
  "portfolio_id" integer,
  "hard_skill_id" integer
);

CREATE TABLE IF NOT EXISTS "portfolio_soft_skills" (
  "id" serial PRIMARY KEY,
  "portfolio_id" integer,
  "soft_skill_id" integer
);

CREATE TABLE IF NOT EXISTS "portfolio_links" (
  "id" serial PRIMARY KEY,
  "link" text,
  "portfolio_id" integer
);

CREATE TABLE IF NOT EXISTS "hard_skills" (
  "id" serial PRIMARY KEY,
  "hard_skill" text
);

CREATE TABLE IF NOT EXISTS "soft_skills" (
  "id" serial PRIMARY KEY,
  "soft_skill" text
);

ALTER TABLE "vacantion_hard_skills" ADD FOREIGN KEY ("vacancy_id") REFERENCES "vacancies" ("id");

ALTER TABLE "vacantion_soft_skills" ADD FOREIGN KEY ("vacancy_id") REFERENCES "vacancies" ("id");

ALTER TABLE "vacantion_hard_skills" ADD FOREIGN KEY ("hard_skill_id") REFERENCES "hard_skills" ("id");

ALTER TABLE "vacantion_soft_skills" ADD FOREIGN KEY ("soft_skill_id") REFERENCES "soft_skills" ("id");

ALTER TABLE "resumes" ADD FOREIGN KEY ("vacancy_id") REFERENCES "vacancies" ("id");

ALTER TABLE "resumes" ADD FOREIGN KEY ("finder_id") REFERENCES "finders" ("id");

ALTER TABLE "finders" ADD FOREIGN KEY ("hr_id") REFERENCES "hr" ("id");

ALTER TABLE "portfolio" ADD FOREIGN KEY ("finder_id") REFERENCES "finders" ("id");

ALTER TABLE "portfolio_hard_skills" ADD FOREIGN KEY ("portfolio_id") REFERENCES "portfolio" ("id");

ALTER TABLE "portfolio_soft_skills" ADD FOREIGN KEY ("portfolio_id") REFERENCES "portfolio" ("id");

ALTER TABLE "portfolio_soft_skills" ADD FOREIGN KEY ("soft_skill_id") REFERENCES "soft_skills" ("id");

ALTER TABLE "portfolio_hard_skills" ADD FOREIGN KEY ("hard_skill_id") REFERENCES "hard_skills" ("id");

ALTER TABLE "portfolio_links" ADD FOREIGN KEY ("portfolio_id") REFERENCES "portfolio" ("id");
