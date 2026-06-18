---
title: Roadmap de producto del TUI
description: Dirección de producto planificada para la interfaz terminal de Stellar Explorer.
---

# Roadmap de producto del TUI

El TUI de Stellar Explorer está planificado como un espacio de trabajo terminal para explorar actividad de Stellar. La dirección del producto se centra en investigación rápida, lecturas con origen visible, claridad Soroban, monitoreo en vivo y contexto local persistente.

Este roadmap describe capacidades futuras de cara al usuario. Está escrito como dirección de producto, no como auditoría.

## Visión del producto

El TUI debe permitir que un usuario empiece desde cualquier objeto de Stellar y recorra su contexto sin salir del terminal. Un flujo típico debería permitir:

1. Abrir el TUI con Stellar RPC directo o con el backend indexado disponible.
2. Buscar un ledger, transacción, cuenta, activo, contrato, etiqueta, marcador o nota.
3. Inspeccionar la entidad en una vista compacta.
4. Saltar a transacciones, operaciones, holders, eventos, contratos, cuentas o activos relacionados.
5. Guardar contexto local durante la investigación.
6. Volver después y recuperar suficiente estado local para continuar.

La base actual ya incluye runtime interactivo, lookup, lecturas con fuente visible, fallback híbrido, metadatos locales, live feed, timelines, specs de contratos, storage de contratos y primeras vistas Soroban decoded/raw. El roadmap describe cómo deben madurar esas superficies.

## Temas del roadmap

| Tema | Resultado esperado |
|---|---|
| Exploración conectada | Las vistas deben sentirse relacionadas, no como consultas aisladas. |
| Entendimiento Soroban | Contratos, eventos, specs, storage e invocaciones deben ser legibles desde terminal. |
| Monitoreo en vivo | La actividad reciente debe evolucionar hacia monitoreo nativo de streams. |
| Workspace local | Marcadores, etiquetas, notas, caché y vistas guardadas deben ser flujos principales. |
| Calidad de búsqueda | La búsqueda debe funcionar bien con input parcial, contexto local y datos indexados. |
| Confianza operativa | Fuente, fallback, caché y estados no disponibles deben ser visibles. |

## Exploración conectada

El TUI debe avanzar desde lookup puntual hacia exploración de relaciones de Stellar. El usuario debe poder abrir una entidad, entender su contexto y seguir entidades relacionadas con pocas teclas.

Implementaciones planificadas:

- Agregar breadcrumbs que muestren cómo se llegó a la vista actual.
- Preservar selección y scroll al entrar y salir de listas anidadas.
- Profundizar transacciones con efectos y mejor contexto de operaciones.
- Agregar vistas dedicadas de detalle de operaciones.
- Hacer consistente el traversal entre ledgers, transacciones, cuentas, activos y contratos.
- Agregar encabezados de contexto para sublistas como transacciones de cuenta, holders de activo y eventos de contrato.

Resultado esperado:

- El usuario puede moverse de un ledger a transacciones, de una transacción a operaciones, de operaciones a cuentas/activos/contratos y volver al contexto original.

## Workspace Soroban

El TUI ya muestra metadatos de contratos, eventos recientes, resúmenes de spec, resúmenes de storage y modos decoded/raw. El siguiente paso es convertir esa base en un workspace de investigación de contratos.

Implementaciones planificadas:

- Agregar vistas de invocaciones con función llamada, argumentos, estado, recursos y contexto de transacción.
- Mostrar autorización cuando el payload lo permita.
- Expandir navegación de métodos más allá de resúmenes compactos.
- Agregar vistas de eventos con secciones de topics, values, decode y raw fallback.
- Agregar drill-down de storage con paginación, expiración, durabilidad, valores decodificados y rutas de copia raw.
- Relacionar contratos con cuentas, activos, transacciones y eventos.

Resultado esperado:

- El usuario puede investigar contratos Stellar desde terminal con contexto decodificado suficiente para uso diario y acceso raw para debugging avanzado.

## Monitoreo en vivo

El live feed actual soporta transacciones recientes, scrollback, pausa/reanudación y filtros generales. El roadmap busca convertirlo en monitoreo real.

Implementaciones planificadas:

- Consumir actualizaciones orientadas a streams desde `services/tui-indexer`.
- Agregar controles de replay sobre actividad retenida.
- Filtrar por cuenta, activo, contrato, clase de transacción, tipo de operación y actividad Soroban.
- Mantener selección estable mientras llega actividad nueva.
- Abrir una fila y volver al mismo contexto de monitoreo.
- Agregar configuración de watch por perfil.

Resultado esperado:

- El usuario puede observar actividad Stellar en vivo, filtrarla, inspeccionar elementos y volver al monitoreo sin perder posición.

## Workspace local

SQLite ya guarda perfiles, sesión, marcadores, etiquetas, notas, scrollback y payloads visitados. El siguiente paso es hacer que ese contexto sea editable y útil dentro de la experiencia principal.

Implementaciones planificadas:

- Crear, editar y eliminar marcadores, etiquetas y notas desde teclado.
- Mostrar metadatos locales en headers, resultados de búsqueda y listas relacionadas.
- Filtrar por entidades marcadas, etiquetadas, anotadas o visitadas.
- Abrir payloads previamente inspeccionados desde caché.
- Crear vistas guardadas para investigaciones recurrentes.
- Aislar workspace por perfil.

Resultado esperado:

- El TUI se convierte en un workspace persistente de investigación, no solo en un cliente de lookup.

## Búsqueda y descubrimiento

La búsqueda ya combina inferencia de entidades, backend y metadatos locales. El siguiente paso es mejorar tolerancia y utilidad ante input parcial o ambiguo.

Implementaciones planificadas:

- Mejorar ranking para cuentas, activos, contratos y transacciones parciales.
- Agrupar resultados por tipo y fuente.
- Mejorar descripciones de resultados indexados y locales.
- Agregar refinamiento o paginación para resultados grandes.
- Buscar contexto local a través de entidades Stellar adjuntas.
- Diferenciar visualmente resultados ejecutables y no ejecutables.

Resultado esperado:

- La búsqueda se vuelve la entrada más rápida al TUI desde hash, dirección, etiqueta, nota, código de activo o identificador parcial.

## Dirección de release

La próxima versión madura debe priorizar:

- navegación conectada entre entidades
- más detalle de transacciones y operaciones
- vistas de invocación y autorización Soroban
- monitoreo nativo de streams
- edición de workspace local
- revisita desde caché
- tests de fallback híbrido, persistencia, búsqueda y render decoded/raw

El producto debe seguir siendo útil en modo RPC directo, mientras que el modo indexado debe aportar exploración más rica, mejor búsqueda y timelines más completos.
