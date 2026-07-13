"use client";

import { QRCodeSVG } from "qrcode.react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { QrCode, Download } from "lucide-react";
import { truncateHash } from "@/lib/utils";
import { useTranslations } from "next-intl";

interface QrDialogProps {
  value: string;
  title?: string;
}

export function QrDialog({ value, title }: QrDialogProps) {
  const t = useTranslations("components.qrDialog");

  const downloadQR = () => {
    const canvas = document.querySelector("#qr-code-svg");
    if (!canvas) return;

    const svgData = new XMLSerializer().serializeToString(canvas);
    const svgBlob = new Blob([svgData], { type: "image/svg+xml;charset=utf-8" });
    const svgUrl = URL.createObjectURL(svgBlob);

    const link = document.createElement("a");
    link.href = svgUrl;
    link.download = `stellar-${truncateHash(value, 8, 0)}.svg`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(svgUrl);
  };

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="ghost" size="icon-xs">
          <QrCode className="size-3.5" />
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-xs">
        <DialogHeader>
          <DialogTitle>{title ?? t("title")}</DialogTitle>
          <DialogDescription className="sr-only">
            Scan this QR code to open the address on a mobile device.
          </DialogDescription>
        </DialogHeader>
        <div className="flex flex-col items-center gap-4 py-4">
          <div className="rounded-lg bg-white p-4">
            <QRCodeSVG id="qr-code-svg" value={value} size={200} level="M" includeMargin={false} />
          </div>
          <p className="text-muted-foreground px-4 text-center font-mono text-xs break-all">
            {value}
          </p>
          <Button variant="outline" size="sm" onClick={downloadQR}>
            <Download className="mr-2 size-4" />
            {t("downloadSvg")}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
