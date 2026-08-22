import { ArrowDown, ArrowUp, ArrowUpDown } from "lucide-react";
import { useMemo, useState, type ReactNode } from "react";
import { Input } from "./Input";

export interface TableColumn<T> {
  id: string;
  header: string;
  cell: (row: T) => ReactNode;
  sortValue?: (row: T) => string | number;
  filterValue?: (row: T) => string;
  className?: string;
}

type SortDirection = "asc" | "desc";

interface TableProps<T> {
  columns: TableColumn<T>[];
  data: T[];
  filterPlaceholder?: string;
  emptyMessage?: string;
  getRowKey: (row: T) => string;
}

export function Table<T>({
  columns,
  data,
  filterPlaceholder = "Filter…",
  emptyMessage = "No records found.",
  getRowKey,
}: TableProps<T>) {
  const [filter, setFilter] = useState("");
  const [sortColumnId, setSortColumnId] = useState<string | null>(null);
  const [sortDirection, setSortDirection] = useState<SortDirection>("asc");

  const filteredAndSorted = useMemo(() => {
    const normalizedFilter = filter.trim().toLowerCase();

    let rows = data;
    if (normalizedFilter) {
      rows = rows.filter((row) =>
        columns.some((column) => {
          if (!column.filterValue) return false;
          return column.filterValue(row).toLowerCase().includes(normalizedFilter);
        })
      );
    }

    if (!sortColumnId) return rows;

    const sortColumn = columns.find((column) => column.id === sortColumnId);
    if (!sortColumn?.sortValue) return rows;

    return [...rows].sort((a, b) => {
      const aValue = sortColumn.sortValue!(a);
      const bValue = sortColumn.sortValue!(b);
      const cmp =
        typeof aValue === "number" && typeof bValue === "number"
          ? aValue - bValue
          : String(aValue).localeCompare(String(bValue));
      return sortDirection === "asc" ? cmp : -cmp;
    });
  }, [columns, data, filter, sortColumnId, sortDirection]);

  function handleSort(columnId: string, sortable: boolean) {
    if (!sortable) return;
    if (sortColumnId === columnId) {
      setSortDirection((current) => (current === "asc" ? "desc" : "asc"));
      return;
    }
    setSortColumnId(columnId);
    setSortDirection("asc");
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="max-w-sm">
        <Input
          label="Search"
          value={filter}
          onChange={(event) => setFilter(event.target.value)}
          placeholder={filterPlaceholder}
        />
      </div>

      <div className="overflow-x-auto rounded border border-ink/10 bg-paper shadow-card">
        <table className="w-full min-w-[640px] border-collapse text-left">
          <thead>
            <tr className="border-b border-ink/10 bg-paper-dim">
              {columns.map((column) => {
                const sortable = Boolean(column.sortValue);
                const isActive = sortColumnId === column.id;
                return (
                  <th
                    key={column.id}
                    scope="col"
                    className={[
                      "px-4 py-3 text-xs font-semibold uppercase tracking-wide text-ink500",
                      sortable ? "cursor-pointer select-none" : "",
                      column.className ?? "",
                    ].join(" ")}
                    onClick={() => handleSort(column.id, sortable)}
                  >
                    <span className="inline-flex items-center gap-1">
                      {column.header}
                      {sortable &&
                        (isActive ? (
                          sortDirection === "asc" ? (
                            <ArrowUp className="h-3.5 w-3.5" aria-hidden />
                          ) : (
                            <ArrowDown className="h-3.5 w-3.5" aria-hidden />
                          )
                        ) : (
                          <ArrowUpDown
                            className="h-3.5 w-3.5 opacity-40"
                            aria-hidden
                          />
                        ))}
                    </span>
                  </th>
                );
              })}
            </tr>
          </thead>
          <tbody>
            {filteredAndSorted.length === 0 ? (
              <tr>
                <td
                  colSpan={columns.length}
                  className="px-4 py-8 text-center text-sm text-ink500"
                >
                  {emptyMessage}
                </td>
              </tr>
            ) : (
              filteredAndSorted.map((row) => (
                <tr
                  key={getRowKey(row)}
                  className="border-b border-ink/5 last:border-b-0 hover:bg-paper-dim/50"
                >
                  {columns.map((column) => (
                    <td
                      key={column.id}
                      className={[
                        "px-4 py-3 text-sm text-ink align-middle",
                        column.className ?? "",
                      ].join(" ")}
                    >
                      {column.cell(row)}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
