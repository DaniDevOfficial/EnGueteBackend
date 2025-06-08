### Database Schema

#### 1. **Users Table**
- Basic table for user information without a dependency on any specific group.
   ```sql
   CREATE TABLE users (
       user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       name VARCHAR(100) NOT NULL,
       email VARCHAR(255) UNIQUE NOT NULL,
       password_hash VARCHAR(255) NOT NULL,
       created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
   );
   ```

#### 2. **Groups Table**
- This table represents a unique group, with `created_by` linking to the user who created the group.
   ```sql
   CREATE TABLE groups (
       group_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       group_name VARCHAR(100) NOT NULL,
       created_by UUID REFERENCES users(user_id) ON DELETE SET NULL,
       created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
   );
   ```

#### 3. **User_Groups Table** (Join table for many-to-many user-group relationship)
- Each record associates a user with a group, allowing users to join multiple groups and groups to have multiple members.
   ```sql
   CREATE TABLE user_groups (
       user_group_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
       group_id UUID REFERENCES groups(group_id) ON DELETE CASCADE,
       joined_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
   );
   ```

#### 4. **Meals Table**
- Represents individual meal entries within a group, with information on type, date, and creation details.
   ```sql
   CREATE TABLE meals (
       meal_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       group_id UUID REFERENCES groups(group_id) ON DELETE CASCADE,
       meal_type VARCHAR(50) NOT NULL, -- e.g., "Lunch" or "Dinner"
       date TIMESTAMPTZ NOT NULL,      -- Date and time of the meal
       title VARCHAR(100) NOT NULL,    -- Title of the meal
       notes TEXT,                     -- Additional notes for the meal
       closed BOOLEAN DEFAULT FALSE,   -- Whether the meal is closed for sign-ups
       fulfilled BOOLEAN DEFAULT FALSE,-- Fulfillment status of the meal
       created_by UUID REFERENCES users(user_id) ON DELETE SET NULL,
       created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
   );
   ```

#### 5. **Meal_Preferences Table**
- Stores meal preferences for each user within a specific meal (e.g., opt-in or opt-out).
- This allows each group member to set their participation for each planned meal.
   ```sql
   CREATE TABLE meal_preferences (
       preference_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       meal_id UUID REFERENCES meals(meal_id) ON DELETE CASCADE,
       user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
       preference VARCHAR(20) NOT NULL, -- e.g., "opt-in", "opt-out"
       created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
   );
   ```

---

### Schema Relationships Summary

- **Users** and **Groups** have a many-to-many relationship via the **User_Groups** table.
- **Meals** are linked to **Groups** with each meal having a `created_by` field to track the creator.
- **Meal_Preferences** links **Users** and **Meals**, allowing flexible participation options for each meal.
- **Meal_Cooks** links **Users** and **Meals** in a many-to-many relationship, allowing multiple users to be assigned as cooks for each meal.

**Diagram Overview:**
- `Users` ← (many-to-many) → `Groups` via `User_Groups`.
- `Groups` → (one-to-many) → `Meals`.
- `Meals` ← (one-to-many) → `Meal_Preferences` ← (many-to-one) → `Users`.
- `Meals` ← (many-to-many) → `Users` via `Meal_Cooks`.