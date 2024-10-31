-- UUID generation (PostgreSQL specific)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Table: users
CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       uuid UUID UNIQUE DEFAULT uuid_generate_v4(),
                       username VARCHAR(255) NOT NULL,
                       email VARCHAR(255) UNIQUE NOT NULL,
                       password VARCHAR(255) NOT NULL,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: tournaments
CREATE TABLE tournaments (
                             id SERIAL PRIMARY KEY,
                             uuid UUID UNIQUE DEFAULT uuid_generate_v4(),
                             name VARCHAR(255) NOT NULL,
                             host_id INT REFERENCES users(id) ON DELETE CASCADE,
                             point_type VARCHAR(50) NOT NULL,
                             scoring_method VARCHAR(50) NOT NULL,
                             can_tie BOOLEAN DEFAULT false,
                             start_date_time TIMESTAMP NOT NULL,
                             is_started BOOLEAN DEFAULT false,
                             created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: teams
CREATE TABLE teams (
                       id SERIAL PRIMARY KEY,
                       uuid UUID UNIQUE DEFAULT uuid_generate_v4(),
                       name VARCHAR(255) NOT NULL,
                       tournament_id INT REFERENCES tournaments(id) ON DELETE CASCADE,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: players
CREATE TABLE players (
                         id SERIAL PRIMARY KEY,
                         uuid UUID UNIQUE DEFAULT uuid_generate_v4(),
                         name VARCHAR(255) NOT NULL,
                         is_registered BOOLEAN DEFAULT false,
                         user_id INT REFERENCES users(id) ON DELETE SET NULL,
                         team_id INT REFERENCES teams(id) ON DELETE CASCADE,
                         tournament_id INT REFERENCES tournaments(id) ON DELETE CASCADE,
                         created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: brackets
CREATE TABLE brackets (
                          id SERIAL PRIMARY KEY,
                          uuid UUID UNIQUE DEFAULT uuid_generate_v4(),
                          tournament_id INT REFERENCES tournaments(id) ON DELETE CASCADE,
                          type VARCHAR(50) NOT NULL,
                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: rounds
CREATE TABLE rounds (
                        id SERIAL PRIMARY KEY,
                        uuid UUID UNIQUE DEFAULT uuid_generate_v4(),
                        bracket_id INT REFERENCES brackets(id) ON DELETE CASCADE,
                        round_number INT NOT NULL,
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: games
CREATE TABLE games (
                       id SERIAL PRIMARY KEY,
                       uuid UUID UNIQUE DEFAULT uuid_generate_v4(),
                       round_id INT REFERENCES rounds(id) ON DELETE CASCADE,
                       team1_id INT REFERENCES teams(id) ON DELETE CASCADE,
                       team2_id INT REFERENCES teams(id) ON DELETE CASCADE,
                       team1_score FLOAT DEFAULT 0,
                       team2_score FLOAT DEFAULT 0,
                       winner_id INT REFERENCES teams(id) ON DELETE SET NULL,
                       is_tie BOOLEAN DEFAULT false,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Mock Data Insertions

-- Users
INSERT INTO users (username, email, password) VALUES
                                                  ('hostuser', 'host@tournament.com', 'passwordhash1'),
                                                  ('player1', 'player1@game.com', 'passwordhash2'),
                                                  ('player2', 'player2@game.com', 'passwordhash3');

-- Tournaments
INSERT INTO tournaments (name, host_id, point_type, scoring_method, can_tie, start_date_time, is_started) VALUES
                                                                                                              ('Summer Tournament', (SELECT id FROM users WHERE username = 'hostuser'), 'points', 'highest', false, '2024-11-10 10:00:00', false),
                                                                                                              ('Winter Tournament', (SELECT id FROM users WHERE username = 'hostuser'), 'time', 'lowest', true, '2024-12-20 14:00:00', false);

-- Teams
INSERT INTO teams (name, tournament_id) VALUES
                                            ('Team Alpha', (SELECT id FROM tournaments WHERE name = 'Summer Tournament')),
                                            ('Team Beta', (SELECT id FROM tournaments WHERE name = 'Summer Tournament')),
                                            ('Team Gamma', (SELECT id FROM tournaments WHERE name = 'Winter Tournament'));

-- Players (anonymous and registered)
INSERT INTO players (name, is_registered, user_id, team_id, tournament_id) VALUES
                                                                               ('Anonymous Player 1', false, NULL, (SELECT id FROM teams WHERE name = 'Team Alpha'), (SELECT id FROM tournaments WHERE name = 'Summer Tournament')),
                                                                               ('player1', true, (SELECT id FROM users WHERE username = 'player1'), (SELECT id FROM teams WHERE name = 'Team Beta'), (SELECT id FROM tournaments WHERE name = 'Summer Tournament')),
                                                                               ('player2', true, (SELECT id FROM users WHERE username = 'player2'), (SELECT id FROM teams WHERE name = 'Team Gamma'), (SELECT id FROM tournaments WHERE name = 'Winter Tournament'));

-- Brackets
INSERT INTO brackets (tournament_id, type) VALUES
                                               ((SELECT id FROM tournaments WHERE name = 'Summer Tournament'), 'single_elimination'),
                                               ((SELECT id FROM tournaments WHERE name = 'Winter Tournament'), 'single_elimination');

-- Rounds
INSERT INTO rounds (bracket_id, round_number) VALUES
                                                  ((SELECT id FROM brackets WHERE tournament_id = (SELECT id FROM tournaments WHERE name = 'Summer Tournament')), 1),
                                                  ((SELECT id FROM brackets WHERE tournament_id = (SELECT id FROM tournaments WHERE name = 'Winter Tournament')), 1);

-- Games (Game 1: Team Alpha vs Team Beta)
INSERT INTO games (round_id, team1_id, team2_id, team1_score, team2_score, winner_id, is_tie) VALUES
    ((SELECT id FROM rounds WHERE bracket_id = (SELECT id FROM brackets WHERE tournament_id = (SELECT id FROM tournaments WHERE name = 'Summer Tournament'))),
     (SELECT id FROM teams WHERE name = 'Team Alpha'), (SELECT id FROM teams WHERE name = 'Team Beta'), 95, 100, (SELECT id FROM teams WHERE name = 'Team Beta'), false);

-- Games (Game 2: Team Gamma vs Bye)
INSERT INTO games (round_id, team1_id, team2_id, team1_score, team2_score, winner_id, is_tie) VALUES
    ((SELECT id FROM rounds WHERE bracket_id = (SELECT id FROM brackets WHERE tournament_id = (SELECT id FROM tournaments WHERE name = 'Winter Tournament'))),
     (SELECT id FROM teams WHERE name = 'Team Gamma'), NULL, 0, 0, (SELECT id FROM teams WHERE name = 'Team Gamma'), false);
