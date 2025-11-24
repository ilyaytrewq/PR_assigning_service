import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '20s', target: 10 },
    { duration: '40s', target: 30 },
    { duration: '20s', target: 0 },
  ],
};

const BASE_URL = 'http://localhost:8080';

export default function () {
  const iter = `${__VU}-${__ITER}`;

  const teamName = `team-${iter}`;
  const teamBody = {
    team_name: teamName,
    members: [
      { user_id: `u-${iter}-1`, username: `user-${iter}-1`, is_active: true },
      { user_id: `u-${iter}-2`, username: `user-${iter}-2`, is_active: true },
      { user_id: `u-${iter}-3`, username: `user-${iter}-3`, is_active: false },
    ],
  };

  let res = http.post(
  `${BASE_URL}/team/add`,
  JSON.stringify(teamBody),
  { headers: { 'Content-Type': 'application/json' } },
  );

  check(res, {
    'create_team: 201 or 400': (r) =>
      r.status === 201 || r.status === 400, 
  });

  const authorId = teamBody.members[0].user_id;

  const prId = `pr-${iter}`;
  res = http.post(
    `${BASE_URL}/pullRequest/create`,
    JSON.stringify({
      pull_request_id: prId,
      pull_request_name: `PR ${iter}`,
      author_id: authorId,
    }),
    { headers: { 'Content-Type': 'application/json' } },
  );

  check(res, {
    'create_pr: 2xx': (r) => r.status >= 200 && r.status < 300,
  });

  sleep(0.2);

  const reviewerId = teamBody.members[1].user_id;
  res = http.get(`${BASE_URL}/users/getReview?user_id=${reviewerId}`);
  check(res, {
    'getReview: 200': (r) => r.status === 200,
  });

  res = http.get(`${BASE_URL}/team/get?team_name=${teamName}`);
  check(res, {
    'getTeam: 200': (r) => r.status === 200,
  });

  res = http.get(`${BASE_URL}/stats`);
  check(res, {
    'stats: 200': (r) => r.status === 200,
  });

}