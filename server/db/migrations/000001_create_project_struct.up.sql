CREATE TABLE "hr" (
  "id" serial PRIMARY KEY,
  "email" text UNIQUE,
  "hash_password" text,
  "username" text UNIQUE NOT NULL
);

CREATE TABLE "vacancies" (
  "id" serial PRIMARY KEY,
  "name" text,
  "hr_id" integer
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

CREATE TABLE "hr_skill_analysis" (
  "id" serial PRIMARY KEY,
  "resume_id" integer NOT NULL,
  "vacancy_id" integer NOT NULL,
  "percent_match" integer,
  "created_at" timestamp DEFAULT now(),
  FOREIGN KEY ("resume_id") REFERENCES "resumes" ("id"),
  FOREIGN KEY ("vacancy_id") REFERENCES "vacancies" ("id")
);

CREATE TABLE "hr_analysis_hard_skills" (
  "id" serial PRIMARY KEY,
  "analysis_id" integer,
  "hard_skill_id" integer,
  "matched" boolean,
  FOREIGN KEY ("analysis_id") REFERENCES "hr_skill_analysis" ("id"),
  FOREIGN KEY ("hard_skill_id") REFERENCES "hard_skills" ("id")
);

CREATE TABLE "hr_analysis_soft_skills" (
  "id" serial PRIMARY KEY,
  "analysis_id" integer,
  "soft_skill_id" integer,
  "matched" boolean,
  FOREIGN KEY ("analysis_id") REFERENCES "hr_skill_analysis" ("id"),
  FOREIGN KEY ("soft_skill_id") REFERENCES "soft_skills" ("id")
);

CREATE TABLE "middle_vacancies" (
  "id" serial PRIMARY KEY,
  "name" text
);

CREATE TABLE "middle_hard_skills" (
  "id" serial PRIMARY KEY,
  "vacancy_id" integer,
  "hard_skill_id" integer
);

CREATE TABLE "middle_soft_skills" (
  "id" serial PRIMARY KEY,
  "vacancy_id" integer,
  "soft_skill_id" integer
);

CREATE TABLE "users" (
  "id" serial PRIMARY KEY,
  "hash_password" text,
  "email" text
);

CREATE TABLE "user_resumes" (
  "id" serial PRIMARY KEY,
  "vacancy_1_id" integer,
  "vacancy_2_id" integer,
  "vacancy_3_id" integer,
  "user_id" integer
);

CREATE TABLE "user_resume_soft" (
  "id" serial PRIMARY KEY,
  "soft_skill_id" integer,
  "resume_id" integer
);

CREATE TABLE "user_resume_hard" (
  "id" serial PRIMARY KEY,
  "hard_skill_id" integer,
  "resume_id" integer
);

CREATE TABLE "user_skill_analysis" (
    "id" serial PRIMARY KEY,
    "resume_id" integer NOT NULL,
    "vacancy_id" integer NOT NULL,
    "percent_match" integer,
    "created_at" timestamp DEFAULT now(),
    FOREIGN KEY ("resume_id") REFERENCES "user_resumes" ("id"),
    FOREIGN KEY ("vacancy_id") REFERENCES "middle_vacancies" ("id")
);

CREATE TABLE "user_analysis_hard_skills" (
    "id" serial PRIMARY KEY,
    "analysis_id" integer,
    "hard_skill_id" integer,
    "matched" boolean,
    FOREIGN KEY ("analysis_id") REFERENCES "user_skill_analysis" ("id"),
    FOREIGN KEY ("hard_skill_id") REFERENCES "hard_skills" ("id")
);

CREATE TABLE "user_analysis_soft_skills" (
    "id" serial PRIMARY KEY,
    "analysis_id" integer,
    "soft_skill_id" integer,
    "matched" boolean,
    FOREIGN KEY ("analysis_id") REFERENCES "user_skill_analysis" ("id"),
    FOREIGN KEY ("soft_skill_id") REFERENCES "soft_skills" ("id")
);

ALTER TABLE "vacantion_hard_skills" ADD FOREIGN KEY ("vacancy_id") REFERENCES "vacancies" ("id");

ALTER TABLE "vacantion_soft_skills" ADD FOREIGN KEY ("vacancy_id") REFERENCES "vacancies" ("id");

ALTER TABLE "vacantion_hard_skills" ADD FOREIGN KEY ("hard_skill_id") REFERENCES "hard_skills" ("id");

ALTER TABLE "vacantion_soft_skills" ADD FOREIGN KEY ("soft_skill_id") REFERENCES "soft_skills" ("id");

ALTER TABLE "resumes" ADD FOREIGN KEY ("vacancy_id") REFERENCES "vacancies" ("id");

ALTER TABLE "resumes" ADD FOREIGN KEY ("finder_id") REFERENCES "finders" ("id");

ALTER TABLE "finders" ADD FOREIGN KEY ("hr_id") REFERENCES "hr" ("id");

ALTER TABLE "vacancies" ADD FOREIGN KEY ("hr_id") REFERENCES "hr" ("id");

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

ALTER TABLE "middle_soft_skills" ADD FOREIGN KEY ("vacancy_id") REFERENCES "middle_vacancies" ("id");

ALTER TABLE "middle_hard_skills" ADD FOREIGN KEY ("vacancy_id") REFERENCES "middle_vacancies" ("id");

ALTER TABLE "middle_hard_skills" ADD FOREIGN KEY ("hard_skill_id") REFERENCES "hard_skills" ("id");

ALTER TABLE "middle_soft_skills" ADD FOREIGN KEY ("soft_skill_id") REFERENCES "soft_skills" ("id");

ALTER TABLE "user_resumes" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "user_resumes" ADD FOREIGN KEY ("vacancy_1_id") REFERENCES "vacancies" ("id");

ALTER TABLE "user_resumes" ADD FOREIGN KEY ("vacancy_2_id") REFERENCES "vacancies" ("id");

ALTER TABLE "user_resumes" ADD FOREIGN KEY ("vacancy_3_id") REFERENCES "vacancies" ("id");

ALTER TABLE "user_resume_hard" ADD FOREIGN KEY ("hard_skill_id") REFERENCES "hard_skills" ("id");

ALTER TABLE "user_resume_soft" ADD FOREIGN KEY ("resume_id") REFERENCES "user_resumes" ("id");

ALTER TABLE "user_resume_hard" ADD FOREIGN KEY ("resume_id") REFERENCES "user_resumes" ("id");

ALTER TABLE "user_resume_soft" ADD FOREIGN KEY ("soft_skill_id") REFERENCES "soft_skills" ("id");