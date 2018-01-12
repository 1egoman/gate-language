export function insert(group) {
  group.append('path')
    .attr('fill', 'transparent')
    .attr('stroke', 'black')
    .attr('stroke-width', 2)

  return group
}

export function merge(group) {
  group.select('path')
    .attr('fill', d => {
      if (d.active) {
        return 'green';
      } else if (d.state === 'on') {
        return 'magenta';
      } else {
        return 'silver';
      }
    })
    .attr('d', d => {
      // 'M' with no circle
      return `M0.5,0.5 L0.5,49.5 L29.5,49.5 L29.5,0.5 L0.5,0.5 Z M14.6385883,7.38237665
      L18.3068605,3.71410446 L23.4263057,3.71410446 L23.4263057,18.4556447
      L19.0622929,18.4556447 L19.0622929,9.2995031 L14.5542,13.807596
      L10.0133185,9.26671443 L10.0133185,18.8044706 L6.5,18.8044706 L6.5,3.5
      L10.7562116,3.5 L14.6385883,7.38237665 Z`
    });

  return group
}
