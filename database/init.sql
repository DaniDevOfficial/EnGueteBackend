CREATE
EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users Table
CREATE TABLE users
(
    user_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR(100)        NOT NULL,
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

-- Group Invites Table
CREATE TABLE group_invites
(
    invite_token UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id     UUID NOT NULL REFERENCES groups (group_id) ON DELETE CASCADE,
    expires_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP
);

-- User_Groups Table (Many-to-Many Relationship between Users and Groups)
CREATE TABLE user_groups
(
    user_group_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
    group_id      UUID NOT NULL REFERENCES groups (group_id) ON DELETE CASCADE,
    joined_at     TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP,

    -- Unique constraint to prevent duplicate user_id, group_id pairs
    CONSTRAINT unique_user_group UNIQUE (user_id, group_id)

);

CREATE TABLE user_groups_blacklist
(
    user_group_blacklist_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                 UUID NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
    group_id                UUID NOT NULL REFERENCES groups (group_id) ON DELETE CASCADE,
    banned_at               TIMESTAMPTZ      DEFAULT CURRENT_TIMESTAMP,

    -- Unique constraint to prevent duplicate user_id, group_id pairs
    CONSTRAINT unique_user_group_blacklist UNIQUE (user_id, group_id)

);

CREATE TABLE user_group_roles
(
    user_group_roles_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_groups_id      UUID        NOT NULL REFERENCES user_groups (user_group_id) ON DELETE CASCADE,
    user_id             UUID        NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
    group_id            UUID        NOT NULL REFERENCES groups (group_id) ON DELETE CASCADE,
    role                VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'manager', 'member')),

    CONSTRAINT unique_user_group_roles UNIQUE (user_groups_id, user_id, group_id, role)
);


-- Meals Table
CREATE TABLE meals
(
    meal_id    UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    group_id   UUID REFERENCES groups (group_id) ON DELETE CASCADE,
    meal_type  VARCHAR(50)  NOT NULL,               -- e.g., "Lunch" or "Dinner"
    date_time  TIMESTAMPTZ  NOT NULL,               -- Date and time of the meal
    title      VARCHAR(100) NOT NULL,               -- Title of the meal
    notes      TEXT,                                -- Additional notes for the meal
    closed     BOOLEAN      NOT NULL DEFAULT FALSE, -- Whether the meal is closed for sign-ups
    fulfilled  BOOLEAN      NOT NULL DEFAULT FALSE, -- Fulfillment status of the meal
    created_by UUID         REFERENCES users (user_id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ           DEFAULT CURRENT_TIMESTAMP
);

-- Meal_Preferences Table (User Preferences for Each Meal)
CREATE TABLE meal_preferences
(
    preference_id UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    meal_id       UUID        NOT NULL REFERENCES meals (meal_id) ON DELETE CASCADE,
    user_id       UUID        NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
    preference    VARCHAR(20) NOT NULL CHECK (preference IN ('opt-in', 'opt-out', 'undecided', 'eat later')),
    changed_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_meal_preference UNIQUE (meal_id, user_id)
);

-- Meal_Cooks Table (Many-to-Many Relationship between Meals and Users)
CREATE TABLE meal_cooks
(
    meal_cook_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    meal_id      UUID NOT NULL REFERENCES meals (meal_id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,

    CONSTRAINT unique_meal_cook UNIQUE (meal_id, user_id)
);
