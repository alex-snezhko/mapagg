export type Response<T> = {
  success: true;
  data: T;
} | {
  success: false;
  error: string;
}

export interface LegendItem {
  color: readonly [number, number, number, number] | null;
  value: number | null;
}

export interface SubmitChoroplethMapRequest {
  tag: string;
  overlayLocTopLeftX: number;
  overlayLocTopLeftY: number;
  overlayLocBottomRightX: number;
  overlayLocBottomRightY: number;
  colorTolerance: number;
  borderTolerance: number;
  legend: LegendItem[];
}

export interface MapAggregation {
  data: [number, number, number][];
  gapY: number;
  gapX: number;
}

export type TagsResponse = Response<string[]>;
export type MapResponse = Response<MapAggregation>;

export interface AggregateDataTag {
  tag: string;
  weight: number;
}

export interface AggregationInputs {
  tags: AggregateDataTag[];
  samplingRate: number;
}
