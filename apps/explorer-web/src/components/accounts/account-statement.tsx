"use client";

import { useState } from "react";
import { Download, FileJson, FileText, FileType2, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { useTranslations } from "next-intl";
import { getHorizonClient } from "@/lib/stellar/client";
import { useNetwork } from "@/lib/providers";
import { getOperationSummary } from "@/lib/utils/operation-summary";
import { asHorizonOperation } from "@/lib/utils/horizon-types";

interface AccountStatementProps {
  accountId: string;
}

type ExportFormat = "json" | "csv" | "pdf";

interface StatementRow {
  date: string;
  ledger: number;
  type: string;
  summary: string;
  txHash: string;
  sourceAccount: string;
}

const MAX_PDF_ROWS = 500;

async function fetchOperationsForAccount(
  accountId: string,
  network: string,
  daysBack: number
): Promise<StatementRow[]> {
  const horizon = getHorizonClient(network as "mainnet" | "testnet" | "futurenet");
  const cutoffDate = new Date(Date.now() - daysBack * 24 * 60 * 60 * 1000);
  const rows: StatementRow[] = [];

  let page = await horizon.operations().forAccount(accountId).order("desc").limit(200).call();

  while (true) {
    for (const op of page.records) {
      const opDate = new Date(op.created_at);
      if (opDate < cutoffDate) return rows;
      rows.push({
        date: op.created_at,
        ledger: parseInt(op.paging_token.split("-")[0]) || 0,
        type: op.type,
        summary: getOperationSummary(asHorizonOperation(op)),
        txHash: op.transaction_hash,
        sourceAccount: op.source_account,
      });
    }
    if (page.records.length < 200) break;
    try {
      page = await page.next();
    } catch {
      break;
    }
  }

  return rows;
}

function downloadFile(content: string, filename: string, mimeType: string) {
  const blob = new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

function toCSV(rows: StatementRow[]): string {
  const headers = ["Date", "Type", "Summary", "Transaction Hash", "Source Account", "Ledger"];
  const lines = [
    headers.join(","),
    ...rows.map((r) =>
      [
        `"${r.date}"`,
        `"${r.type}"`,
        `"${r.summary.replace(/"/g, '""')}"`,
        `"${r.txHash}"`,
        `"${r.sourceAccount}"`,
        r.ledger,
      ].join(",")
    ),
  ];
  return lines.join("\n");
}

async function loadLogoDataUrl(): Promise<string | null> {
  try {
    const origin = typeof window !== "undefined" ? window.location.origin : "";
    const res = await fetch(`${origin}/stellar-explorer.png`);
    if (!res.ok) return null;
    const blob = await res.blob();
    return await new Promise<string>((resolve, reject) => {
      const reader = new FileReader();
      reader.onloadend = () => resolve(reader.result as string);
      reader.onerror = reject;
      reader.readAsDataURL(blob);
    });
  } catch {
    return null;
  }
}

async function exportPDF(
  accountId: string,
  network: string,
  daysBack: number,
  allRows: StatementRow[]
) {
  const { jsPDF } = await import("jspdf");

  const rows = allRows.slice(0, MAX_PDF_ROWS);
  const truncated = allRows.length > MAX_PDF_ROWS;
  const origin = typeof window !== "undefined" ? window.location.origin : "";
  const logoDataUrl = await loadLogoDataUrl();
  const doc = new jsPDF({ orientation: "portrait", unit: "mm", format: "a4" });

  const PW = 210;
  const M = 14;
  const CW = PW - M * 2;

  // ── Header ──
  doc.setFillColor(250, 250, 250);
  doc.rect(0, 0, PW, 30, "F");

  const lx = M;
  const ly = 8;
  const logoSize = 9;

  if (logoDataUrl) {
    doc.addImage(logoDataUrl, "PNG", lx, ly, logoSize, logoSize);
  }

  doc.setFont("helvetica", "bold");
  doc.setFontSize(13);
  doc.setTextColor(10, 10, 10);
  doc.text("StellarView Explorer", lx + (logoDataUrl ? logoSize + 2 : 0), ly + 6);

  doc.setFont("helvetica", "normal");
  doc.setFontSize(8);
  doc.setTextColor(100, 100, 100);
  doc.text("Account Statement", lx + (logoDataUrl ? logoSize + 2 : 0), ly + 11);

  // Network + date top right
  const netLabel =
    network === "mainnet" ? "Mainnet" : network.charAt(0).toUpperCase() + network.slice(1);
  doc.setTextColor(80, 80, 80);
  doc.setFontSize(8);
  doc.text(
    `${netLabel}  ·  ${new Date().toLocaleDateString("en-US", { year: "numeric", month: "long", day: "numeric" })}`,
    PW - M,
    ly + 6,
    { align: "right" }
  );
  doc.text(`Last ${daysBack} days`, PW - M, ly + 11, { align: "right" });

  // Divider
  doc.setDrawColor(220, 220, 220);
  doc.setLineWidth(0.3);
  doc.line(M, 30, PW - M, 30);

  // ── Account block ──
  let y = 38;
  doc.setFont("helvetica", "normal");
  doc.setFontSize(7.5);
  doc.setTextColor(120, 120, 120);
  doc.text("ACCOUNT", M, y);
  y += 4;
  doc.setFont("helvetica", "bold");
  doc.setFontSize(8.5);
  doc.setTextColor(10, 10, 10);
  doc.text(accountId, M, y, { maxWidth: CW });
  y += 5;

  doc.setFont("helvetica", "normal");
  doc.setFontSize(7.5);
  doc.setTextColor(120, 120, 120);
  doc.text(
    `${rows.length}${truncated ? ` of ${allRows.length}` : ""} operations · Period: last ${daysBack} days${truncated ? ` · Showing first ${MAX_PDF_ROWS}` : ""}`,
    M,
    y
  );
  y += 8;

  doc.setDrawColor(220, 220, 220);
  doc.line(M, y, PW - M, y);
  y += 5;

  // ── Column headers ──
  const COL_DATE = M;
  const COL_TYPE = M + 22;
  const COL_SUMMARY = M + 48;
  const COL_HASH = M + 122;

  doc.setFontSize(7);
  doc.setFont("helvetica", "bold");
  doc.setTextColor(140, 140, 140);
  doc.text("DATE / TIME", COL_DATE, y);
  doc.text("TYPE", COL_TYPE, y);
  doc.text("DESCRIPTION", COL_SUMMARY, y);
  doc.text("TRANSACTION", COL_HASH, y);
  y += 2;

  doc.setDrawColor(200, 200, 200);
  doc.line(M, y, PW - M, y);
  y += 4;

  // ── Rows ──
  const ROW_H = 8;
  let rowIdx = 0;
  doc.setFont("helvetica", "normal");

  for (const row of rows) {
    if (y + ROW_H > 280) {
      doc.addPage();
      doc.setFillColor(250, 250, 250);
      doc.rect(0, 0, PW, 10, "F");

      y = 14;
      doc.setFontSize(7);
      doc.setFont("helvetica", "bold");
      doc.setTextColor(140, 140, 140);
      doc.text("DATE / TIME", COL_DATE, y);
      doc.text("TYPE", COL_TYPE, y);
      doc.text("DESCRIPTION", COL_SUMMARY, y);
      doc.text("TRANSACTION", COL_HASH, y);
      y += 2;
      doc.setDrawColor(200, 200, 200);
      doc.line(M, y, PW - M, y);
      y += 4;
      doc.setFont("helvetica", "normal");
    }

    // Zebra stripe
    if (rowIdx % 2 === 0) {
      doc.setFillColor(247, 248, 250);
      doc.rect(M, y - 3, CW, ROW_H, "F");
    }

    const d = new Date(row.date);
    const dateStr = d.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
    const timeStr = `${String(d.getUTCHours()).padStart(2, "0")}:${String(d.getUTCMinutes()).padStart(2, "0")} UTC`;

    doc.setFontSize(7);
    doc.setTextColor(30, 30, 30);
    doc.text(dateStr, COL_DATE, y + 1);
    doc.setTextColor(140, 140, 140);
    doc.text(timeStr, COL_DATE, y + 4.5);

    // Type label
    doc.setTextColor(60, 60, 60);
    doc.setFont("helvetica", "bold");
    const typeLabel = row.type.replace(/_/g, " ").slice(0, 14);
    doc.text(typeLabel, COL_TYPE, y + 1);
    doc.setFont("helvetica", "normal");

    // Summary
    doc.setTextColor(30, 30, 30);
    const summaryMaxW = COL_HASH - COL_SUMMARY - 4;
    const lines = doc.splitTextToSize(row.summary, summaryMaxW);
    doc.text(lines[0], COL_SUMMARY, y + 1);
    if (lines[1]) {
      doc.setTextColor(100, 100, 100);
      doc.text(lines[1], COL_SUMMARY, y + 4.5);
    }

    // TX hash as clickable link
    const shortHash = `${row.txHash.slice(0, 8)}…${row.txHash.slice(-6)}`;
    const txUrl = `${origin}/en/${network}/tx/${row.txHash}`;
    doc.setTextColor(37, 99, 235);
    doc.text(shortHash, COL_HASH, y + 1);
    doc.link(COL_HASH, y - 3, CW - (COL_HASH - M), ROW_H, { url: txUrl });

    y += ROW_H;
    rowIdx++;
  }

  // ── Footer ──
  const totalPages = doc.getNumberOfPages();
  for (let p = 1; p <= totalPages; p++) {
    doc.setPage(p);
    doc.setDrawColor(220, 220, 220);
    doc.line(M, 287, PW - M, 287);
    doc.setFontSize(7);
    doc.setFont("helvetica", "normal");
    doc.setTextColor(160, 160, 160);
    doc.text("Data sourced from Stellar Horizon API. For informational purposes only.", M, 292);
    doc.text(`Page ${p} / ${totalPages}`, PW - M, 292, { align: "right" });
  }

  const shortId = `${accountId.slice(0, 6)}-${accountId.slice(-4)}`;
  doc.save(`stellar-statement-${shortId}-${new Date().toISOString().slice(0, 10)}.pdf`);
}

export function AccountStatement({ accountId }: AccountStatementProps) {
  const { network } = useNetwork();
  const t = useTranslations("account.statement");
  const [open, setOpen] = useState(false);
  const [daysBack, setDaysBack] = useState(30);
  const [selectedFormat, setSelectedFormat] = useState<ExportFormat>("pdf");
  const [loading, setLoading] = useState(false);
  const [count, setCount] = useState<number | null>(null);

  const handleExport = async () => {
    setLoading(true);
    setCount(null);
    try {
      const rows = await fetchOperationsForAccount(accountId, network, daysBack);
      setCount(rows.length);
      const date = new Date().toISOString().slice(0, 10);
      const shortId = `${accountId.slice(0, 6)}-${accountId.slice(-4)}`;

      if (selectedFormat === "json") {
        const payload = {
          account: accountId,
          network,
          exportedAt: new Date().toISOString(),
          periodDays: daysBack,
          totalOperations: rows.length,
          operations: rows,
        };
        downloadFile(
          JSON.stringify(payload, null, 2),
          `stellar-statement-${shortId}-${date}.json`,
          "application/json"
        );
      } else if (selectedFormat === "csv") {
        downloadFile(toCSV(rows), `stellar-statement-${shortId}-${date}.csv`, "text/csv");
      } else {
        await exportPDF(accountId, network, daysBack, rows);
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm" className="gap-1.5">
          <Download className="size-3.5" />
          {t("exportButton")}
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{t("dialogTitle")}</DialogTitle>
          <DialogDescription className="sr-only">
            Download your account transaction history as JSON, CSV, or PDF.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2">
          <div className="space-y-2">
            <p className="text-muted-foreground text-sm">{t("period")}</p>
            <div className="flex gap-2">
              {[7, 30, 90].map((days) => (
                <Button
                  key={days}
                  variant={daysBack === days ? "default" : "outline"}
                  size="sm"
                  className="flex-1"
                  onClick={() => setDaysBack(days)}
                >
                  {days}d
                </Button>
              ))}
            </div>
          </div>

          <div className="space-y-2">
            <p className="text-muted-foreground text-sm">{t("format")}</p>
            <div className="grid grid-cols-3 gap-2">
              <Button
                onClick={() => setSelectedFormat("json")}
                variant={selectedFormat === "json" ? "default" : "outline"}
                className="h-auto flex-col gap-1 py-3"
              >
                <FileJson className="size-4" />
                <span className="text-xs">JSON</span>
              </Button>
              <Button
                onClick={() => setSelectedFormat("csv")}
                variant={selectedFormat === "csv" ? "default" : "outline"}
                className="h-auto flex-col gap-1 py-3"
              >
                <FileText className="size-4" />
                <span className="text-xs">CSV</span>
              </Button>
              <Button
                onClick={() => setSelectedFormat("pdf")}
                variant={selectedFormat === "pdf" ? "default" : "outline"}
                className="h-auto flex-col gap-1 py-3"
              >
                <FileType2 className="size-4" />
                <span className="text-xs">PDF</span>
              </Button>
            </div>
          </div>

          {count !== null && !loading && (
            <p className="text-muted-foreground text-center text-sm">
              {t("opsCount", { count, days: daysBack })}
              {count > MAX_PDF_ROWS && (
                <span className="text-warning block text-xs">
                  PDF limited to first {MAX_PDF_ROWS}
                </span>
              )}
            </p>
          )}

          {loading && (
            <div className="flex items-center justify-center gap-2 py-2">
              <Loader2 className="size-4 animate-spin" />
              <span className="text-muted-foreground text-sm">Fetching operations…</span>
            </div>
          )}

          <p className="text-muted-foreground text-center text-xs">
            Data from Stellar Horizon · Tx hashes link to explorer
          </p>

          <Button onClick={handleExport} disabled={loading} className="w-full gap-2">
            {loading ? (
              <Loader2 className="size-4 animate-spin" />
            ) : (
              <Download className="size-4" />
            )}
            {t("downloadButton")}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
