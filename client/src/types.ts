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

export interface LatLong {
	lat: number;
	long: number;
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

export interface SubmitPointsOfInterestFromCsvData {
	tag: string;
	minThresholdRadiusMiles: number;
	maxThresholdRadiusMiles: number;
	latCol: string;
	longCol: string;
  weightCol: string | null;
}

export interface PointOfInterest {
  latLong: LatLong;
  weight: number;
}

export interface PointOfInterestWithId extends PointOfInterest {
  id: number;
}

export interface SubmitPointsOfInterestData {
	tag: string;
	pointsOfInterest: PointOfInterest[];
	minThresholdRadiusMiles: number;
	maxThresholdRadiusMiles: number;
}

export type LatLongValue = [number, number, number];

export interface ComponentData {
  tag: string;
  data: number[][];
}

export interface MapAggregation {
  aggregateData: LatLongValue[];
  componentsData: ComponentData[];
  gapY: number;
  gapX: number;
}

export interface AggregateDataTag {
  tag: string;
  weight: number;
}

export interface AggregationInputs {
  tags: AggregateDataTag[];
  samplingRate: number;
}

export interface OverlayBounds {
	topLeft: LatLong;
	bottomRight: LatLong;
}

export type TagsResponse = Response<string[]>;
export type MapResponse = Response<MapAggregation>;
export type OverlayBoundsResponse = Response<OverlayBounds>;
