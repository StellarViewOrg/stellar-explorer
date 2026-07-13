"use client";

import { useEffect, useRef, useState } from "react";
import { geoOrthographic, geoPath, geoGraticule, geoContains, geoBounds } from "d3-geo";

const LAND_URL =
  "https://cdn.jsdelivr.net/gh/martynafford/natural-earth-geojson@master/110m/physical/ne_110m_land.json";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type LandGeoJSON = { type: "FeatureCollection"; features: any[] };

export function NetworkGlobe() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [land, setLand] = useState<LandGeoJSON | null>(null);
  const [useFallback, setUseFallback] = useState(false);

  const rafRef = useRef<number | null>(null);
  const phiRef = useRef(0);
  const allDotsRef = useRef<[number, number][]>([]);
  const isFallbackRef = useRef(false);

  const dragRef = useRef<{ startX: number; startPhi: number } | null>(null);
  const autoRotateRef = useRef(true);
  const resumeTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // ── Fetch GeoJSON ────────────────────────────────────────────────────────
  useEffect(() => {
    let alive = true;
    fetch(LAND_URL)
      .then((r) => {
        if (!r.ok) throw new Error("fetch failed");
        return r.json() as Promise<LandGeoJSON>;
      })
      .then((d) => {
        if (alive) setLand(d);
      })
      .catch(() => {
        if (alive) setUseFallback(true);
      });
    return () => {
      alive = false;
    };
  }, []);

  // ── Pre-generate land dots when GeoJSON arrives ──────────────────────────
  useEffect(() => {
    if (!land) return;
    const step = 1.5;
    const dots: [number, number][] = [];
    for (const feature of land.features) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const [[minLng, minLat], [maxLng, maxLat]] = geoBounds(feature as any);
      for (let lng = minLng; lng <= maxLng; lng += step) {
        for (let lat = minLat; lat <= maxLat; lat += step) {
          if (geoContains(feature, [lng, lat])) {
            dots.push([lng, lat]);
          }
        }
      }
    }
    allDotsRef.current = dots;
  }, [land]);

  // ── Update fallback ref ──────────────────────────────────────────────────
  useEffect(() => {
    isFallbackRef.current = useFallback;
  }, [useFallback]);

  // ── Animation loop (runs once on mount) ─────────────────────────────────
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const rawCtx = canvas.getContext("2d");
    if (!rawCtx) return;
    const ctx: CanvasRenderingContext2D = rawCtx;

    const dpr = Math.min(window.devicePixelRatio, 2);
    const sz = canvas.offsetWidth || 420;
    canvas.width = sz * dpr;
    canvas.height = sz * dpr;
    ctx.scale(dpr, dpr);

    const cx = sz / 2;
    const cy = sz / 2;
    const R = sz * 0.46;

    const projection = geoOrthographic().scale(R).translate([cx, cy]).clipAngle(90);

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const pg = geoPath(projection as any, ctx);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const graticule: any = geoGraticule().step([30, 30])();

    function drawFrame() {
      const fallback = isFallbackRef.current;

      ctx.clearRect(0, 0, sz, sz);

      if (autoRotateRef.current) phiRef.current += 0.25;
      projection.rotate([phiRef.current, -15, 0]);

      // 1. Globe rim
      ctx.beginPath();
      ctx.arc(cx, cy, R, 0, Math.PI * 2);
      ctx.strokeStyle = "#ffffff";
      ctx.lineWidth = 1.5;
      ctx.stroke();

      // 2. Graticule
      ctx.beginPath();
      pg(graticule);
      ctx.strokeStyle = "#ffffff";
      ctx.lineWidth = 0.5;
      ctx.globalAlpha = fallback ? 0.4 : 0.2;
      ctx.stroke();
      ctx.globalAlpha = 1;

      // 3. Land dots — all dots (front + back hemisphere for see-through effect)
      if (allDotsRef.current.length > 0) {
        ctx.fillStyle = "#999999";
        for (const [lng, lat] of allDotsRef.current) {
          const xy = projection([lng, lat]);
          if (!xy) continue;
          ctx.beginPath();
          ctx.arc(xy[0], xy[1], 1.2, 0, Math.PI * 2);
          ctx.fill();
        }
      }

      rafRef.current = requestAnimationFrame(drawFrame);
    }

    rafRef.current = requestAnimationFrame(drawFrame);
    return () => {
      if (rafRef.current) cancelAnimationFrame(rafRef.current);
    };
  }, []);

  // ── Drag interaction ─────────────────────────────────────────────────────
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const onPointerDown = (e: PointerEvent) => {
      e.preventDefault();
      canvas.setPointerCapture(e.pointerId);
      dragRef.current = { startX: e.clientX, startPhi: phiRef.current };
      autoRotateRef.current = false;
      if (resumeTimerRef.current) clearTimeout(resumeTimerRef.current);
    };

    const onPointerMove = (e: PointerEvent) => {
      if (!dragRef.current) return;
      const dx = e.clientX - dragRef.current.startX;
      phiRef.current = dragRef.current.startPhi + dx * 0.35;
    };

    const onPointerUp = () => {
      dragRef.current = null;
      resumeTimerRef.current = setTimeout(() => {
        autoRotateRef.current = true;
      }, 600);
    };

    canvas.addEventListener("pointerdown", onPointerDown);
    canvas.addEventListener("pointermove", onPointerMove);
    canvas.addEventListener("pointerup", onPointerUp);
    canvas.addEventListener("pointercancel", onPointerUp);

    return () => {
      canvas.removeEventListener("pointerdown", onPointerDown);
      canvas.removeEventListener("pointermove", onPointerMove);
      canvas.removeEventListener("pointerup", onPointerUp);
      canvas.removeEventListener("pointercancel", onPointerUp);
      if (resumeTimerRef.current) clearTimeout(resumeTimerRef.current);
    };
  }, []);

  return (
    <div className="relative flex items-center justify-center">
      <canvas
        ref={canvasRef}
        className="w-full cursor-grab active:cursor-grabbing"
        style={{ aspectRatio: "1/1" }}
      />
    </div>
  );
}
