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
      } else if (d.State === 'on') {
        return 'magenta';
      } else {
        return 'silver';
      }
    })
    .attr('d', d => {
      if (d.State === 'on') {
        return `M0.5,0.5 L0.5,49.5 L29.5,49.5 L29.5,0.5 L0.5,0.5 Z M9,14.5 C5.96243388,14.5
        3.5,12.0375661 3.5,9 C3.5,5.96243388 5.96243388,3.5 9,3.5 C12.0375661,3.5
        14.5,5.96243388 14.5,9 C14.5,12.0375661 12.0375661,14.5 9,14.5 Z`;
      } else {
        // Circle on bottom
        return `M0.5,0.5 L0.5,49.5 L29.5,49.5 L29.5,0.5 L0.5,0.5 Z M21,14.5
        C17.9624339,14.5 15.5,12.0375661 15.5,9 C15.5,5.96243388 17.9624339,3.5 21,3.5
        C24.0375661,3.5 26.5,5.96243388 26.5,9 C26.5,12.0375661 24.0375661,14.5 21,14.5 Z`
      }
    });

  return group
}
