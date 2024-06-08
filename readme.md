# years: Powerful Time-Based Navigation in [Go](https://go.dev/)

![Years Logo](https://raw.githubusercontent.com/amberpixels/years/main/years.png "Years")

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/amberpixels/years/blob/main/LICENSE)

`years` is a Golang library designed for time-based sorting and iterating over things. <br>
`years` allows you to build powerful tools for navigating through time-based data structures (calendars files, logs
etc).<br> Also it's yet another time parser for Go as well.

## Features

- **Calendar Navigation**: Provides easy navigation through calendar-like structures, allowing for relative and absolute
  time-based navigation.
- **Flexible Time Representation**: Supports simple date strings, Unix timestamps, and files (flat or nested
  structures).
- **File-Based Time Retrieval**: Handles time based on file metadata (modification/creation/access time) or file names.
- **Time Parsing and Manipulation**: Offers powerful time parsing capabilities and time mutation functions.

## Installation

To install the `years` library, use the following `go get` command:

```sh
go get github.com/amberpixels/years
```

## Usage

### Basic Example

Here's a simple example demonstrating how to use years with strings representing dates:

```go
dates := []string{"2020, Dec 1", "2020, Dec 2", "2020, Dec 3"}
layout := "2006, Jan 2" // Go-formatted time layout

v := years.NewVoyager(years.WaypointGroupFromStrings(dates), layout)

// iterates through all dates in the Future->Past direction
// Here w.Identifier() returns the string value itself
v.Traverse(func (w Waypoint) { fmt.Println(w.Identifier()) }, years.O_PAST())

// Dates are not required to be same layout:
dates2 := []string{"2020-01-01", "2020-01", "2020-Jan-03"}
years.ExtendParserDefaults(years.WithLayouts("2006-01-02", "2006-01-02", "2006-Jan-02"))
v2 := years.NewVoyager(years.WaypointGroupFromStrings(dates2)) // not specifying layouts, so default are used
v2.Traverse(func (w Waypoint) { fmt.Println(w.Identifier()) }, years.O_PAST())

```

### Advanced Example with Files

The following example shows how to work with a nested file structure representing calendar dates:

```go
// Declaring path to a calendar directory
var CalendarPath = "path/to/calendar"
const layout = "2006/Jan/2006-01-02.txt" // using Golang time package layout

wf, err := years.NewTimeNamedWaypointFile(CalendarPath, layout)
if err != nil {
    panic(err)
}

v = years.NewVoyager(wf)
// iterates through all finite files (excluding directories) in Past->Future direction
v.Traverse(func (w Waypoint) {
    fmt.Println(w.Path())
}, years.O_FUTURE(), years.O_LEAVES_ONLY())

// Quick navigation through the calendar
yesterday, err := v.Navigate("yesterday")
if err != nil {
    panic(err)
}
fmt.Println("Yesterday's file:", w.Path())
```

## Time Parsing and Manipulation

`years` can also be used as a time-handling library. It provides various time parsing and mutation functions:

### Parsing time

```go
// 1. Simplest case: parses time almost the same way as Go's time.Parse
t, err := years.Parse("2006-01-02", "2024-05-26")
if err != nil {
    panic(err)
}
fmt.Println("Parsed time:", t)

// Note: Difference is in the fact that it supports layouts with timestamp parts:
// e.g. `U@` for second timestamps, `U@000` for millisecond timestamps, etc
t, err = years.Parse("logs-U@000.log", "logs-1717852417000.log")
if err != nil {
    panic(err)
}
fmt.Println("Parsed time:", t)

// 2. Advanced parsing:
p := NewParser(
    AcceptUnixSeconds(),
    AcceptAliases(),
    WithLayouts(
        "2006",
        "2006-01", "2006-Jan",
        "2006-Jan-02", "2006-01-02",
    ),
)

t, _ = p.Parse("", "2020-01") // not specifying layouts will use all parser's accepted layouts
t, _ = p.JustParse("2020-01") // syntax sugar

// aliases:
t, _ = p.JustParse("today")
t, _ = p.JustParse("next-week")
// etc

// 3. Configuring global parser:
years.SetParserDefaults(
    AcceptUnixSeconds(),
    AcceptAliases(),
)

t, _ = years.JustParse("1717852417")

```

### Mutating time

```go
t, _ := time.Parse("now")
mutatedTime := years.Mutate(&t).TruncateToDay().Time()
fmt.Println("Mutated time:", mutatedTime)
```

## Contributing
`years` welcomes contributions! Feel free to open issues, suggest improvements, or submit pull
requests. [Contribution guidelines for this project](CONTRIBUTING.md)

## License
This project is [licensed under the MIT License](LICENSE).

## Contact
For any questions or issues, feel free to open an issue on the GitHub repository or contact the maintainer at eugene@amberpixels.io
