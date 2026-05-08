export const asset = (name: string) => `/assets/${name}`;

export const buildingImageByBid: Record<number, string> = {
  5: "CityInnerPanel_house.png",
  6: "CityInnerPanel_cityhall.png",
  7: "CityInnerPanel_institute.png",
  8: "CityInnerPanel_forceyard.png",
  9: "CityInnerPanel_barrack.png",
  10: "CityInnerPanel_inn.png",
  11: "CityInnerPanel_recruithall.png",
  12: "CityInnerPanel_temple.png",
  13: "CityInnerPanel_market.png",
  14: "CityInnerPanel_blacksmith.png",
  15: "CityInnerPanel_warhouse.png",
  16: "CityInnerPanel_stable.png",
  17: "CityInnerPanel_barn.png",
  18: "CityInnerPanel_hotel.png",
  19: "CityInnerPanel_tower.png",
  20: "CityInnerPanel_townwall.png"
};

export const buildingIntroByBid: Record<number, string> = {
  1: "building_intro_1.png",
  2: "building_intro_2.png",
  3: "building_intro_3.png",
  4: "building_intro_4.png",
  5: "building_intro_5.png",
  6: "building_intro_6.png",
  7: "building_intro_7.png",
  8: "building_intro_8.png",
  9: "building_intro_9.png",
  10: "building_intro_10.png",
  11: "building_intro_11.png",
  12: "building_intro_12.png",
  13: "building_intro_13.png",
  14: "building_intro_14.png",
  15: "building_intro_15.png",
  16: "building_intro_16.png",
  17: "building_intro_17.png",
  18: "building_intro_18.png",
  19: "building_intro_19.png",
  20: "building_intro_20.png"
};

export const provinceAssets = [
  { id: 1, name: "司隶", image: "silv.png", x: 208.5, y: 159, w: 179, h: 93 },
  { id: 2, name: "冀州", image: "jizhou.png", x: 361, y: 90, w: 102, h: 105 },
  { id: 3, name: "豫州", image: "yuzhou.png", x: 331, y: 190, w: 123, h: 101 },
  { id: 4, name: "兖州", image: "yunzhou.png", x: 374, y: 155, w: 96, h: 92 },
  { id: 5, name: "徐州", image: "xuzhou.png", x: 421, y: 183, w: 118, h: 96 },
  { id: 6, name: "青州", image: "qingzhou.png", x: 428, y: 121, w: 119, h: 81 },
  { id: 7, name: "荆州", image: "jingzhou.png", x: 254, y: 235, w: 152, h: 172 },
  { id: 8, name: "扬州", image: "yangzhou.png", x: 373, y: 239, w: 204, h: 175 },
  { id: 9, name: "益州", image: "yizhou.png", x: 40, y: 220, w: 270, h: 204 },
  { id: 10, name: "凉州", image: "liangzhou.png", x: 0, y: 7, w: 260, h: 244 },
  { id: 11, name: "并州", image: "bingzhou.png", x: 219, y: 45, w: 153, h: 155 },
  { id: 12, name: "幽州", image: "youzhou.png", x: 360, y: 4, w: 253, h: 118 },
  { id: 13, name: "交州", image: "jiaozhou.png", x: 135, y: 363, w: 328, h: 136 }
];

export type Plot = {
  position: number;
  gridX: number;
  gridY: number;
  x: number;
  y: number;
  w: number;
  h: number;
};

export const INNER_GRID_LEFT = 356;
export const INNER_GRID_TOP = 130;
export const INNER_GRID_OFFX = 54;
export const INNER_GRID_OFFY = 27;
export const INNER_GRID_COUNT = 6;

export function isInnerCityHallGrid(gridX: number, gridY: number) {
  return (gridX === 2 || gridX === 3) && (gridY === 0 || gridY === 1);
}

export function isInnerWallGrid(gridX: number, gridY: number) {
  return (
    (gridX < 0 || gridX > INNER_GRID_COUNT - 1 || gridY < 0 || gridY > INNER_GRID_COUNT - 1) &&
    gridX > -3 &&
    gridX < INNER_GRID_COUNT + 2 &&
    gridY > -3 &&
    gridY < INNER_GRID_COUNT + 2
  );
}

export function innerGridPoint(x: number, y: number) {
  return {
    gridX: Math.floor((2 * (y - INNER_GRID_TOP) + (x - INNER_GRID_LEFT)) / (INNER_GRID_OFFX * 2)),
    gridY: Math.floor((2 * (y - INNER_GRID_TOP) - (x - INNER_GRID_LEFT)) / (INNER_GRID_OFFX * 2))
  };
}

export function innerPosition(gridX: number, gridY: number) {
  return 100 + gridX * 10 + gridY;
}

export function innerGridRect(gridX: number, gridY: number): Plot {
  return {
    position: innerPosition(gridX, gridY),
    gridX,
    gridY,
    x: INNER_GRID_LEFT + (gridX - gridY) * INNER_GRID_OFFX - INNER_GRID_OFFX,
    y: INNER_GRID_TOP + (gridX + gridY) * INNER_GRID_OFFY - 25,
    w: 108,
    h: 79
  };
}

export const innerPlots: Plot[] = Array.from({ length: INNER_GRID_COUNT * INNER_GRID_COUNT }, (_, index) => {
  const gridX = index % INNER_GRID_COUNT;
  const gridY = Math.floor(index / INNER_GRID_COUNT);
  return { gridX, gridY };
})
  .filter(({ gridX, gridY }) => !isInnerCityHallGrid(gridX, gridY))
  .map(({ gridX, gridY }) => innerGridRect(gridX, gridY));

export const cityHallPlot: Plot = {
  position: 81,
  gridX: 2,
  gridY: 0,
  x: 368,
  y: 182,
  w: 195,
  h: 109
};

export const cityWallPlot: Plot = {
  position: 199,
  gridX: 9,
  gridY: 9,
  x: 0,
  y: 24,
  w: 731,
  h: 519
};

const OUTER_GRID_LEFT = 230;
const OUTER_GRID_TOP = 135;
const OUTER_GRID_OFFX = 54;
const OUTER_GRID_OFFY = 27;
const OUTER_GRID_COUNT_X = 9;
const OUTER_GRID_COUNT_Y = 7;
const OUTER_GOVERNMENT_LEVEL_MAX = 10;

export const outerValidGridByGovernmentLevel: number[][] = [
  [0,1,0,0,0,0,1,0,0,0,0,1,1,1,1,1,1,0,0,0,0,0,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
  [0,1,0,0,0,0,1,1,0,0,0,1,1,1,1,1,1,1,0,0,0,0,1,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
  [0,1,0,0,0,0,1,1,0,0,0,1,1,1,1,1,1,1,0,0,0,0,1,1,1,1,1,0,0,0,0,0,0,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
  [0,1,0,0,0,0,1,1,0,0,0,1,1,1,1,1,1,1,1,0,0,0,1,1,1,1,1,1,0,0,0,0,0,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
  [0,1,0,0,0,0,1,1,0,0,1,1,1,1,1,1,1,1,1,0,0,1,1,1,1,1,1,1,0,0,0,0,1,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
  [0,1,0,0,0,0,1,1,0,0,1,1,1,1,1,1,1,1,1,0,0,1,1,1,1,1,1,1,0,0,0,0,1,1,1,1,1,0,0,0,0,0,0,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
  [0,1,0,0,0,0,1,1,0,0,1,1,1,1,1,1,1,1,1,0,0,1,1,1,1,1,1,1,1,0,0,0,1,1,1,1,1,1,0,0,0,0,0,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
  [0,1,0,0,0,0,1,1,0,0,1,1,1,1,1,1,1,1,1,0,0,1,1,1,1,1,1,1,1,0,0,0,1,1,1,1,1,1,0,0,0,0,0,1,1,1,1,0,0,0,0,0,0,0,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0],
  [0,1,0,0,0,0,1,1,0,0,1,1,1,1,1,1,1,1,1,0,1,1,1,1,1,1,1,1,1,0,0,1,1,1,1,1,1,1,0,0,0,0,1,1,1,1,1,0,0,0,0,0,0,0,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0],
  [0,1,0,0,0,0,1,1,0,0,1,1,1,1,1,1,1,1,1,0,1,1,1,1,1,1,1,1,1,0,0,1,1,1,1,1,1,1,0,0,0,0,1,1,1,1,1,0,0,0,0,0,0,1,1,1,1,0,0,0,0,0,0,0,1,1,0,0,0,0]
];

export function isOuterCityWallGrid(gridX: number, gridY: number) {
  return gridX >= 2 && gridX <= 5 && gridY >= -3 && gridY <= 0;
}

export function outerPosition(gridX: number, gridY: number) {
  return gridY * 10 + gridX;
}

export function outerGridRect(gridX: number, gridY: number): Plot {
  return {
    position: outerPosition(gridX, gridY),
    gridX,
    gridY,
    x: OUTER_GRID_LEFT + (gridX - gridY) * OUTER_GRID_OFFX - OUTER_GRID_OFFX,
    y: OUTER_GRID_TOP + (gridX + gridY) * OUTER_GRID_OFFY - 12,
    w: 112,
    h: 66
  };
}

export const outerPlots: Plot[] = Array.from({ length: OUTER_GRID_COUNT_X * OUTER_GRID_COUNT_Y }, (_, index) => {
  const gridX = index % OUTER_GRID_COUNT_X;
  const gridY = Math.floor(index / OUTER_GRID_COUNT_X);
  return outerGridRect(gridX, gridY);
});

export function outerPlotsForGovernmentLevel(level: number): Plot[] {
  const validLevel = Math.min(Math.max(Math.floor(level) || 1, 1), OUTER_GOVERNMENT_LEVEL_MAX);
  const validGrid = outerValidGridByGovernmentLevel[validLevel - 1];
  return outerPlots.filter(
    (plot) => !isOuterCityWallGrid(plot.gridX, plot.gridY) && validGrid[plot.gridY * 10 + plot.gridX] === 1
  );
}
