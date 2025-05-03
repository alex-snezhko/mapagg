<template>
  <div class="about">
    <h1>MapAgg</h1>

    <p>
      MapAgg is a map aggregation tool designed to help visualize how different user-determined factors relate to different locations within a geographic area.
    </p>

    <section>
      <h2>Instructions</h2>

      <ol>
        <li>
          Begin by selecting the rectangular boundaries of the geographic area you are interested in, and optionally an overlay to define the land area of interest within the rectangular boundary.
        </li>
        <li>
          Navigate to the 'Add Map Dataset' page to add different datasets of interest relating to the selected area. The goal of each map is to pinpoint which locations within your geographic area
          are 'better' or 'worse' with respect to a variable of interest.
        </li>
        <li>
          Navigate to the home page to visualize how all of the previously-selected datasets come together to form an aggregate map. Areas colored in green indicate a better performance for
          the selected metrics and areas in red indicate a worse performance. Each factor can be assigned a different weight, which will be used in the overall calculation of the score for
          each location on the map.
        </li>
      </ol>
    </section>

    <section>
      <h2>Example Usage</h2>

      <p>I am interested in moving to a new location. I have several factors that I value and want to see how different areas compare against each other.</p>

      <ol>
        <li>
          I am interesting in moving to a location within the USA. I select the boundaries of the USA and for my overlay input a map image like this:

          <img class="example-overlay-img" src="/assets/usa_land_area_overlay.jpg" />
        </li>
        <li>
          One of the factors I value is housing affordability. I look online for a choropleth map comparing average home prices in each county, and select to add a new choropleth map dataset
          according to this map. I assign higher values to cheaper counties and lower values to more expensive counties, and confirm my dataset.
        </li>
        <li>
          I also value distance to my family and friends, who live in various locations within the country. I am OK travelling long distance to visit them but would prefer a closer location if
          possible. I select to add a new points of interest proximity map dataset, and input the locations of my friends and family along with radial distances that I consider feasible for
          travel.
        </li>
        <li>
          I continue adding other maps for factors I care about, and when I'm done I go to the aggregation home page. I value some factors more than others so assign different weights to each
          factor, and then see a visualization of which locations are better and which are worse.
        </li>
      </ol>
    </section>
  </div>
</template>

<style lang="scss" scoped>
.about {
  width: 1100px;
  margin: auto;
  padding: 30px;

  h1 {
    font-weight: bold;
    text-align: center;
  }

  h2 {
    font-weight: bold;
  }

  section {
    margin: 30px 0;
  }

  p {
    margin: 20px 0;
  }
}

.example-overlay-img {
  width: 300px;
}
</style>
