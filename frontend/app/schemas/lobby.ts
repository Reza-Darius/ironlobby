import { z } from "zod";
import {
  CalendarDate,
  Time,
  toCalendarDateTime,
  now,
  getLocalTimeZone,
} from "@internationalized/date";

export const gameModes = ["Vanilla", "modded", "rp"] as const;

export const lobbySchema = z
  .object({
    title: z.string().min(3, "Title must be at least 3 characters"),
    description: z.string().max(500, "Description should be max 500 characters"),
    date: z.instanceof(CalendarDate, { message: "Please select a Date" }),
    time: z.instanceof(Time, { message: "Please select a Time" }),
    countries: z
      .array(
        z.object({
          tag: z.string().min(1, "Please select a country"),
          players: z.number().int().min(1, "At least 1 player"),
        }),
      )
      .min(1, "At least 1 country")
      .refine(
        (arr) => new Set(arr.map((c) => c.tag)).size === arr.length,
        "Länder dürfen sich nicht wiederholen",
      ),
    gameMode: z.enum(gameModes, "Please select a game mode"),
  })
  .refine(
    ({ date, time }) =>
      toCalendarDateTime(date, time).compare(toCalendarDateTime(now(getLocalTimeZone()))) > 0,
    { message: "Date must be in the future", path: ["date"] },
  );

export type LobbySchema = z.output<typeof lobbySchema>;
