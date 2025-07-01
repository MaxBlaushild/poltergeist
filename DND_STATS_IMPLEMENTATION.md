# D&D Stats System Implementation

This document outlines the implementation of a D&D-style stats system for users, including automatic stat point allocation when leveling up.

## Overview

The system adds six classic D&D stats (Strength, Dexterity, Constitution, Intelligence, Wisdom, Charisma) to users and automatically awards stat points that can be allocated when users level up.

## Components Implemented

### 1. User Stats Model (`go/pkg/models/user_stats.go`)

- **UserStats struct**: Contains the six D&D stats with default value of 10 each
- **Constants**:
  - `DefaultStatValue = 10`: Starting value for all stats
  - `StatPointsPerLevel = 5`: Number of stat points awarded per level up
  - `MaxStatValue = 20`: Maximum value for any stat
  - `MinStatValue = 8`: Minimum value for any stat (not enforced in current implementation)

- **Key Methods**:
  - `AddStatPoints(points int)`: Adds available stat points when leveling up
  - `AllocateStatPoint(statName string)`: Allocates one stat point to a specific stat
  - `GetStatValue(statName string)`: Gets the value of a specific stat
  - `GetStatModifier(statName string)`: Calculates D&D-style modifiers ((value - 10) / 2)
  - `GetAllStats()`: Returns a map of all stat values
  - `GetAllStatModifiers()`: Returns a map of all stat modifiers

### 2. Database Migration (`go/migrate/internal/migrations/000095_create_user_stats.*`)

- **Up migration**: Creates `user_stats` table with:
  - All six D&D stats (default 10)
  - `available_stat_points` field (default 0)
  - Foreign key to users table
  - Standard timestamps and UUID

- **Down migration**: Drops the `user_stats` table

### 3. Database Handler (`go/pkg/db/user_stats.go`)

- **userStatsHandler**: Provides database operations for user stats
- **Methods**:
  - `FindOrCreateForUser()`: Creates stats with default values if they don't exist
  - `FindByUserID()`: Retrieves stats for a specific user
  - `Create()` and `Update()`: Basic CRUD operations
  - `AllocateStatPoint()`: Handles stat point allocation with validation
  - `AddStatPoints()`: Adds stat points when user levels up

### 4. Database Interface (`go/pkg/db/interfaces.go`)

- **UserStatsHandle interface**: Defines the contract for user stats operations
- Added to main `DbClient` interface
- Integrated into the database client factory

### 5. Level Up Integration (`go/sonar/internal/gameengine/client.go`)

- **Enhanced `awardExperiencePoints()`**: Now automatically awards stat points when users level up
- **Process**: 
  1. User gains experience and levels up
  2. System calculates stat points to award (levels gained × 5)
  3. Automatically adds stat points to user's available pool
  4. User can then allocate these points via API

### 6. API Endpoints (`go/sonar/internal/server/server.go`)

#### GET `/sonar/stats`
- **Purpose**: Retrieve user's current stats and available stat points
- **Authentication**: Required
- **Response**: JSON with all stats, modifiers, and available points

#### POST `/sonar/stats/allocate`
- **Purpose**: Allocate one stat point to a specific stat
- **Authentication**: Required
- **Request Body**: `{"statName": "strength|dexterity|constitution|intelligence|wisdom|charisma"}`
- **Response**: Updated stats with success message
- **Validation**: 
  - Ensures stat name is valid
  - Checks available stat points > 0
  - Prevents stats from exceeding maximum value (20)

## Usage Examples

### Getting User Stats
```bash
GET /sonar/stats
Authorization: Bearer <token>

Response:
{
  "id": "uuid",
  "strength": 12,
  "dexterity": 10,
  "constitution": 14,
  "intelligence": 10,
  "wisdom": 10,
  "charisma": 8,
  "availableStatPoints": 3,
  "modifiers": {
    "strength": 1,
    "dexterity": 0,
    "constitution": 2,
    "intelligence": 0,
    "wisdom": 0,
    "charisma": -1
  }
}
```

### Allocating Stat Points
```bash
POST /sonar/stats/allocate
Authorization: Bearer <token>
Content-Type: application/json

{
  "statName": "strength"
}

Response:
{
  "strength": 13,
  "availableStatPoints": 2,
  "message": "strength increased to 13",
  ...
}
```

## System Flow

1. **User completes quest/challenge** → Gains experience points
2. **Experience processing** → User levels up (via existing system)
3. **Automatic stat point award** → 5 stat points added per level gained
4. **User allocation** → Player uses API to allocate points to desired stats
5. **Stat modifiers** → Used for gameplay mechanics (future implementation)

## Configuration

- **Stat Points per Level**: Currently set to 5, configurable via `StatPointsPerLevel` constant
- **Maximum Stat Value**: Set to 20 (D&D standard), configurable via `MaxStatValue` constant
- **Default Stat Values**: All stats start at 10 (D&D average), configurable via `DefaultStatValue` constant

## Database Schema

```sql
CREATE TABLE user_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL,
    strength INTEGER NOT NULL DEFAULT 10,
    dexterity INTEGER NOT NULL DEFAULT 10,
    constitution INTEGER NOT NULL DEFAULT 10,
    intelligence INTEGER NOT NULL DEFAULT 10,
    wisdom INTEGER NOT NULL DEFAULT 10,
    charisma INTEGER NOT NULL DEFAULT 10,
    available_stat_points INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

## Future Enhancements

- **Stat-based gameplay mechanics**: Use stat modifiers for challenge difficulty
- **Equipment bonuses**: Items that modify stats temporarily
- **Stat requirements**: Certain actions/items requiring minimum stat values
- **Respec functionality**: Allow users to reset and reallocate their stats
- **Stat caps by level**: Prevent over-allocation at low levels
- **Achievement integration**: Award bonus stat points for specific achievements

## Notes

- The system integrates seamlessly with the existing level/experience system
- All new user stats are initialized with default values when first accessed
- The API includes both raw stat values and calculated D&D modifiers
- Error handling includes validation for invalid stat names and insufficient points
- The migration is numbered 000095 to fit the existing migration sequence