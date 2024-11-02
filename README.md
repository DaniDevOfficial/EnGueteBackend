# En Guete Backend

This is the backend for the *En Guete* app, designed to allow users to manage meal groups, opt in or out of meals, and
invite others to join meal groups.

---

## Use Cases

### User

1. [x] As a user, I want to create and manage my account.
2. [ ] As a user, I want to create and join multiple groups.
3. [ ] As a user, I want to invite other users to my groups via an invite link.
4. [ ] As a user, I want to view all my groups and see which users belong to each group.
5. [ ] As a user, I want to leave any group I am a member of.

### Group Admin (Creator of a Group)

1. [ ] As a group admin, I want to create meal plans for my group.
2. [ ] As a group admin, I want to update or delete my group.
3. [ ] As a group admin, I want to see the meal preferences of all members for a given meal.
4. [ ] As a group admin, I want to remove members from my group if needed.

---

## Backend API Endpoints

### User-Related Endpoints

| **Use Case**          | **Endpoint**              | **Method**       | **Implemented** |
|-----------------------|---------------------------|------------------|-----------------|
| Create/Manage Account | `/users`                  | GET, PUT, DELETE | done            |
| Sign in               | `/auth/signin`            | POST             | done            |
| Sign up               | `/auth/signup`            | POST             | done            |
| View all groups       | `/users/:user_id/groups`  | GET              | partially done  |
| Leave a group         | `/groups/:group_id/leave` | DELETE           | -               |

### Group-Related Endpoints

| **Use Case**                           | **Endpoint**                       | **Method** | **Implemented** |
|----------------------------------------|------------------------------------|------------|-----------------|
| Create a new group                     | `/groups`                          | POST       | done            |
| Invite a user to join a group (link)   | `/groups/:group_id/invite`         | GET        | done            |
| View all members of a group            | `/groups/:group_id/members`        | GET        | -               |
| Delete group (admin only)              | `/groups/:group_id`                | DELETE     | -               |
| Create a meal plan (admin only)        | `/groups/:group_id/meals`          | POST       | -               |
| View meal preferences of group members | `/groups/:group_id/meals/:meal_id` | GET        | -               |

---

## Data Models

### User Model

```typescript
import {UUID} from "some-uuid-library";

type User = {
    user_id: UUID;
    name: string;
    email: string;
    password_hash: string;
    created_at: Date;
};
```

### Group Model

```typescript
type Group = {
    group_id: UUID;
    group_name: string;
    created_by: UUID; // References the User who created the group
    created_at: Date;
};
```

### User_Groups Model (for Many-to-Many Relationship)

```typescript
type UserGroup = {
    user_group_id: UUID;
    user_id: UUID;      // References User
    group_id: UUID;     // References Group
    joined_at: Date;
};
```

### Meal Model

```typescript
type Meal = {
    meal_id: UUID;
    group_id: UUID;     // References Group
    meal_type: string;  // e.g., "Lunch", "Dinner"
    date: Date;         // The date the meal is planned for
    created_by: UUID;   // References the admin or creator
    created_at: Date;
};
```

### Meal_Preferences Model (for storing user preferences per meal)

```typescript
type MealPreference = {
    preference_id: UUID;
    meal_id: UUID;      // References Meal
    user_id: UUID;      // References User
    preference: string; // e.g., "opt-in", "opt-out", or dietary preferences
};
```

---

## Group Management and Invitation Logic

1. **Group Creation**:
    - Users can create a new group, automatically making them the group admin.
    - On creation, a record is added to the `User_Groups` table, associating the admin with the new group.

2. **Invitation Process**:
    - The group admin can generate an invitation link unique to their group.
    - The link includes the `group_id`, which allows the backend to connect the invitee to the correct group.
    - When an invitee follows the link, they can register or log in, after which they are added to the `User_Groups`
      table for that group.

3. **Group Membership**:
    - Users can view all the groups they are part of.
    - Users can leave groups, removing their association in the `User_Groups` table.

4. **Meal Planning and Preferences**:
    - Group admins can create meal entries specifying the meal type and date.
    - Members can select their preferences (opt-in/opt-out) for each meal.
    - Group admins can view meal preferences for planning purposes.

---

## Example JSON Responses

### Group Invitation Response

```json
{
  "group_id": "123e4567-e89b-12d3-a456-426614174000",
  "invite_link": "https://enguete.app/invite/123e4567-e89b-12d3-a456-426614174000"
}
```

### Meal Preferences Response

```json
{
  "meal_id": "323e4567-e89b-12d3-a456-426614174001",
  "date": "2024-11-01",
  "preferences": [
    {
      "user_id": "123e4567-e89b-12d3-a456-426614174002",
      "preference": "opt-in"
    },
    {
      "user_id": "223e4567-e89b-12d3-a456-426614174003",
      "preference": "opt-out"
    }
  ]
}
```

---

### Meal Planning Workflow

1. **Admin Creates Meal**: Admin specifies meal type, date, and any notes.
2. **User Preferences**: Members mark preferences (opt-in or out) for that meal.
3. **View Preferences**: Admin can see the list of preferences for planning.
