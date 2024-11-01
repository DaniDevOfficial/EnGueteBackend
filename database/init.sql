CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users Table

CREATE TABLE users
(
    user_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(100)        NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255)        NOT NULL,
    created_at    TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP
);

-- Groups Table
CREATE TABLE groups
(
    group_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_name VARCHAR(100) NOT NULL,
    created_by UUID         REFERENCES users (user_id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP
);

-- User_Groups Table (Many-to-Many Relationship between Users and Groups)
CREATE TABLE user_groups
(
    user_group_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID REFERENCES users (user_id) ON DELETE CASCADE,
    group_id      UUID REFERENCES groups (group_id) ON DELETE CASCADE,
    joined_at     TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP
);

-- Meals Table
CREATE TABLE meals
(
    meal_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id   UUID REFERENCES groups (group_id) ON DELETE CASCADE,
    meal_type  VARCHAR(50)  NOT NULL,          -- e.g., "Lunch" or "Dinner"
    dateTime  TIMESTAMPTZ  NOT NULL,          -- Date and time of the meal
    title      VARCHAR(100) NOT NULL,          -- Title of the meal
    notes      TEXT,                           -- Additional notes for the meal
    closed     BOOLEAN      DEFAULT FALSE,     -- Whether the meal is closed for sign-ups
    fulfilled  BOOLEAN          DEFAULT FALSE, -- Fulfillment status of the meal
    created_by UUID         REFERENCES users (user_id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP
);

-- Meal_Preferences Table (User Preferences for Each Meal)
CREATE TABLE meal_preferences
(
    preference_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    meal_id       UUID REFERENCES meals (meal_id) ON DELETE CASCADE,
    user_id       UUID REFERENCES users (user_id) ON DELETE CASCADE,
    preference    VARCHAR(20) NOT NULL, -- e.g., "opt-in", "opt-out", "undecided", "eat later"
    created_at    TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP
);

-- Meal_Cooks Table (Many-to-Many Relationship between Meals and Users)
CREATE TABLE meal_cooks
(
    meal_cook_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    meal_id      UUID REFERENCES meals (meal_id) ON DELETE CASCADE,
    user_id      UUID REFERENCES users (user_id) ON DELETE CASCADE
);
