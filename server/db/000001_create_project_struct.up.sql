CREATE TABLE "hr" (
  "id" serial PRIMARY KEY,
  "email" text UNIQUE,
  "hash_password" text,
  "username" text UNIQUE NOT NULL
);

CREATE TABLE "vacancies" (
  "id" serial PRIMARY KEY,
  "name" text
);

CREATE TABLE "vacantion_hard_skills" (
  "id" serial PRIMARY KEY,
  "vacancy_id" integer,
  "hard_skill_id" integer
);

CREATE TABLE "vacantion_soft_skills" (
  "id" serial PRIMARY KEY,
  "vacancy_id" integer,
  "soft_skill_id" integer
);

CREATE TABLE "finders" (
  "id" serial PRIMARY KEY,
  "portfolio" boolean,
  "hr_id" integer
);

CREATE TABLE "resumes" (
  "id" serial PRIMARY KEY,
  "first_name" text,
  "last_name" text,
  "surname" text,
  "phone_number" text,
  "email" text,
  "vacancy_id" integer,
  "finder_id" integer
);

CREATE TABLE "resume_hard_skill" (
  "id" serial PRIMARY KEY,
  "resume_id" integer,
  "hard_skill_id" integer
);

CREATE TABLE "resume_soft_skill" (
  "id" serial PRIMARY KEY,
  "resume_id" integer,
  "soft_skill_id" integer
);

CREATE TABLE "portfolio" (
  "id" serial PRIMARY KEY,
  "finder_id" integer
);

CREATE TABLE "portfolio_hard_skills" (
  "id" serial PRIMARY KEY,
  "portfolio_id" integer,
  "hard_skill_id" integer
);

CREATE TABLE "portfolio_soft_skills" (
  "id" serial PRIMARY KEY,
  "portfolio_id" integer,
  "soft_skill_id" integer
);

CREATE TABLE "portfolio_links" (
  "id" serial PRIMARY KEY,
  "link" text,
  "portfolio_id" integer
);

CREATE TABLE "hard_skills" (
  "id" serial PRIMARY KEY,
  "hard_skill" text
);

CREATE TABLE "soft_skills" (
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

ALTER TABLE "resume_soft_skill" ADD FOREIGN KEY ("resume_id") REFERENCES "resumes" ("id");

ALTER TABLE "resume_hard_skill" ADD FOREIGN KEY ("resume_id") REFERENCES "resumes" ("id");

ALTER TABLE "resume_hard_skill" ADD FOREIGN KEY ("hard_skill_id") REFERENCES "hard_skills" ("id");

ALTER TABLE "resume_soft_skill" ADD FOREIGN KEY ("soft_skill_id") REFERENCES "soft_skills" ("id");
