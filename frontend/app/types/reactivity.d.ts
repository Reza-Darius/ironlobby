import type { CalendarDate, CalendarDateTime, Time, ZonedDateTime } from "@internationalized/date";

declare module "@vue/reactivity" {
  export interface RefUnwrapBailTypes {
    internationalizedDate: CalendarDate | CalendarDateTime | Time | ZonedDateTime;
  }
}
