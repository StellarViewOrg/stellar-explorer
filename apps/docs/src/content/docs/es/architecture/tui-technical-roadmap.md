---
title: Roadmap técnico del TUI
description: Líneas técnicas planificadas para la interfaz terminal de Stellar Explorer.
---

# Roadmap técnico del TUI

Este roadmap describe el trabajo técnico planificado para el TUI de Stellar Explorer. Está organizado por líneas de implementación para poder planificar cambios futuros sin convertir la documentación en una auditoría.

La base actual incluye:

- runtime Bubble Tea y shell Lip Gloss
- modos RPC, indexado e híbrido
- metadatos de fuente y etiquetas de fallback
- lookup de ledgers, transacciones, cuentas, activos y contratos
- listas, sublistas y timelines indexados
- SQLite local para perfiles, metadatos, scrollback y entidades visitadas
- primeras vistas Soroban decoded/raw para eventos, specs y storage

## Línea 1: Navegación y composición

Propósito:

- hacer que la exploración terminal se sienta continua a medida que las vistas sean más profundas

Implementación planificada:

- Agregar estado de breadcrumbs al modelo de la app.
- Preservar selección y scroll en transiciones anidadas.
- Agregar headers de contexto para sublistas.
- Normalizar acciones de entidades relacionadas para que cada fila tenga comando, etiqueta y valor copiable.
- Separar vistas complejas en secciones reutilizables manteniendo el ownership de teclado en cada modelo.

Consideraciones técnicas:

- Los breadcrumbs deben derivarse de acciones de navegación, no duplicarse manualmente en render.
- El estado local de vistas debe seguir dentro de modelos Bubble Tea.
- `LookupExplorerSnapshot` puede seguir como puente de sublistas, pero vistas más profundas pueden requerir un descriptor de ruta más rico.

## Línea 2: Transacciones y operaciones

Propósito:

- convertir el lookup de transacciones en una superficie completa de investigación

Implementación planificada:

- Agregar efectos a detalle de transacciones cuando haya datos indexados.
- Agregar vistas de detalle de operación con campos específicos por tipo.
- Conectar operaciones con cuentas source, destinos, activos, contratos y transacciones padre.
- Mejorar resúmenes de operaciones Soroban y actividad de activos.
- Agregar tests de traversal transacción-operación y selección de comandos.

Consideraciones técnicas:

- Los efectos deben venir de APIs indexadas cuando sea posible.
- El modo RPC directo debe comunicar claramente secciones no disponibles.
- Los renderers de operaciones deben preferir campos tipados sobre resúmenes de texto.

## Línea 3: Inspección Soroban

Propósito:

- hacer que contratos y actividad Soroban sean comprensibles desde terminal

Implementación planificada:

- Agregar modelos de detalle de invocación.
- Mostrar función, argumentos, estado, recursos y contexto de transacción.
- Agregar vistas de autorización cuando el payload lo soporte.
- Expandir eventos más allá de resúmenes compactos hacia topics y values.
- Agregar paginación y comandos de drill-down para storage.
- Mantener modos decoded y raw en cada sección Soroban relevante.

Consideraciones técnicas:

- Las vistas decodificadas nunca deben ocultar los valores raw.
- Storage y eventos deben usar el patrón existente de `limit` y `offset`.
- El parsing de specs debe tolerar specs parciales o no disponibles.

## Línea 4: Monitoreo en vivo

Propósito:

- evolucionar live feed desde actividad reciente refrescada hacia monitoreo nativo de streams

Implementación planificada:

- Agregar consumidor de streams para canales de `services/tui-indexer`.
- Mantener polling como fallback.
- Agregar replay sobre scrollback retenido.
- Agregar filtros por cuenta, activo, contrato, tipo de operación, clase de transacción y actividad Soroban.
- Preservar fila seleccionada y viewport mientras llegan filas nuevas.
- Agregar tests de orden, deduplicación, pausa y estabilidad de selección.

Consideraciones técnicas:

- El live feed debe mantener scrollback limitado.
- La ingestión de stream debe deduplicar por hash y preservar orden de ledger/aplicación.
- Pause debe evitar mutación visible sin perder eventos necesarios para replay.

## Línea 5: Workspace local y caché

Propósito:

- hacer que el estado local sea útil para investigación real

Implementación planificada:

- Crear, editar y eliminar marcadores, etiquetas y notas.
- Agregar filtros de metadatos locales en búsqueda y lookup.
- Abrir entidades inspeccionadas desde caché.
- Usar caché como fallback cuando backend/RPC no estén disponibles.
- Agregar vistas guardadas con comando, filtros, entidad y perfil.
- Agregar tests de migración por cada cambio de esquema SQLite.

Consideraciones técnicas:

- Payloads cacheados deben incluir fuente, timestamp, resumen y metadatos tipados suficientes.
- El fallback de caché debe ser explícito en la metadata de fuente.
- Las vistas guardadas deben referenciar comandos y filtros, evitando duplicar estado visual.

## Línea 6: Búsqueda y ranking

Propósito:

- hacer que la búsqueda sea la entrada más rápida hacia datos Stellar y contexto local

Implementación planificada:

- Mejorar ranking local y backend para input parcial.
- Agrupar resultados por tipo y fuente.
- Agregar paginación o refinamiento para resultados grandes.
- Hacer ejecutables resultados de etiquetas, marcadores, notas y caché.
- Agregar tests de ranking, dedupe, filas deshabilitadas y metadata adjunta.

Consideraciones técnicas:

- El ranking debe preferir matches exactos, luego prefijos, luego metadata.
- Los resultados locales deben seguir disponibles sin backend.
- Los límites del backend deben ser explícitos para mantener la paleta rápida.

## Línea 7: Confiabilidad y release

Propósito:

- mantener el producto predecible a medida que crecen rutas de lectura y estado local

Implementación planificada:

- Expandir tests de fallback híbrido y resolución de una sola fuente.
- Agregar tests de timelines, paginación y render Soroban decoded/raw.
- Cubrir migraciones de SQLite.
- Agregar tests de read API para efectos, invocaciones, storage y eventos.
- Agregar fixtures para estados degradados.

Consideraciones técnicas:

- Tests que requieren Stellar RPC, data lake o PostgreSQL local deben separarse de unit tests puros.
- Tests con red deben documentar o saltar sus prerequisitos cuando la infraestructura local no esté disponible.
- La documentación debe mantenerse alineada con el comportamiento implementado después de cada hito.

## Línea 8: Documentación y alineación de release

Propósito:

- mantener docs de arquitectura, guías de setup y workflows de CI alineados con la alpha implementada después de cada hito

Implementación planificada:

- Actualizar docs de arquitectura y setup para reflejar capacidades y limitaciones actuales del TUI.
- Documentar tiers de tests y prerequisitos de infraestructura para `apps/tui` y `services/tui-indexer`.
- Agregar workflows de CI para build, lint, unit tests e integración del terminal.
- Extender tests de read API para live feed, búsqueda y subrecursos Soroban paginados.

Consideraciones técnicas:

- La documentación debe describir dirección de capacidades sin convertir el roadmap en una auditoría de release.
- CI debe ejecutar tiers rápidos y deterministas por defecto, dejando checks con red como opcionales o documentados.
- Las páginas en inglés y español deben mantenerse sincronizadas para arquitectura y setup de desarrollo.

## Orden recomendado

1. Navegación y composición.
2. Profundidad de transacciones y operaciones.
3. Invocaciones y autorización Soroban.
4. Monitoreo nativo de streams.
5. Workspace local editable y apertura desde caché.
6. Ranking y agrupación de búsqueda.
7. Cobertura de confiabilidad.
8. Documentación y alineación de release.

Este orden mantiene el TUI útil en cada paso mientras aumenta la profundidad esperada de un explorador terminal profesional para Stellar.
