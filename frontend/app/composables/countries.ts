interface Country {
  country_tag: string;
  country_name: string;
}

export function useCountries() {
  return useFetch("/api/country", {
    key: "countries",
    baseURL: "http://localhost:8000",
    transform: (data: Country[]) =>
      data.map((c) => ({ label: c.country_name, value: c.country_tag })),
    default: () => [],
  });
}
