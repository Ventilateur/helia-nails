# Booking platforms synchronisation

## Supported platforms

- Treatwell Connect
- Planity
- Classpass (via Google Calendar integration)

## Untracked AWS resources

- S3 bucket for Terraform state
- Config file (same S3 bucket)
- Secret file (Parameter Store)

## How to add a new employee (aka. new calendar)

A new employee needs to be added on all platforms.

First, create a new employee entry in config file:

```yaml
employees:
  - name: "Kimmy"
    treatwell:
      id: ${TREATWELL_ID}
    planity:
      id: ${PLANITY_ID}
    classpass:
      googleCalendarId: ${GOOGLE_CALENDAR_ID}
```

### Classpass (Google Calendar)

- Create a new agenda with the employee's name.
- In the calendar's parameter, share the calendar with the Google Cloud service account. It needs at least `Edit Events`.
- Replace `${GOOGLE_CALENDAR_ID}` with the calendar ID.

### Treatwell

- Create a new employee on Treatwell Connect.
- Go on Treatwell Connect, open the developer tool and search for a request named `employees.json`.
- Find the employee's id under `employees[].id` and replace `${TREATWELL_ID}` with it.

### Planity

- Create a new employee on Planity.
- Go on Planity Pro, open the developer tool and book a fake appointment with note `<employee name>_TEST`.
- Filter network for `WS` (websocket) and search for messages that contain `<employee name>_TEST`.
- There will be a JSON message with a key or value that looks like `calendar_vevents/<some weird ass id>`.
- Replace `${PLANITY_ID}` with that `<some weird ass id>`.

Now push the config file to S3, the next run will update the new employee's calendar.
